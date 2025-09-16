package database

import (
	"cmp"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)


type DynamicDB struct {
	Database
	Ents  []Entity
	Repos map[string]*Collection[Entity]
}

type Entity interface {
	Table() string
	GetModel() *Model
}

func Dynamic(engine Database, opts ...DynamicDBOption) *DynamicDB {
	db := DynamicDB{engine, []Entity{}, map[string]*Collection[Entity]{}}
	for _, opt := range opts {
		opt(&db)
	}
	return &db
}

type DynamicDBOption func(*DynamicDB)

func WithModel(ent Entity) DynamicDBOption {
	return func(db *DynamicDB) {
		if err := db.Register(ent); err != nil {
			// Log the error but don't exit - the application can decide how to handle it
			log.Printf("WARNING: failed to register model %s: %v", ent.Table(), err)
		}
	}
}

func (db *DynamicDB) Model() Model {
	return Model{DB: db}
}

func (db *DynamicDB) NewModel(id string) Model {
	return Model{db, id, time.Now(), time.Now()}
}

func (db *DynamicDB) Register(ent Entity) error {
	if err := db.Query(`
		CREATE TABLE IF NOT EXISTS ` + ent.Table() + ` (
			ID        TEXT PRIMARY KEY,
			CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`).Exec(); err != nil {
		return errors.Wrap(err, "failed to create table")
	}

	kind := reflect.ValueOf(ent).Kind()
	if kind != reflect.Ptr && kind != reflect.Struct {
		return errors.New("expected struct, got " + kind.String())
	}

	fields, types, defaults := db.Fields(ent)
	for i, field := range fields {
		db.Query(fmt.Sprintf(`
			ALTER TABLE %s ADD COLUMN %s %s DEFAULT %v
		`, ent.Table(), field, types[i], defaults[i])).Exec()
	}

	db.Ents = append(db.Ents, ent)
	return nil
}

func (db *DynamicDB) Fields(ent Entity) (fields []string, types []string, defaults []string) {
	value := reflect.ValueOf(ent)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return
	}

	type_ := value.Type()
	for i := range type_.NumField() {
		field := type_.Field(i)
		kind := field.Type.Kind()

		// Skip anonymous fields, pointers, interfaces, and functions
		if field.Anonymous || kind == reflect.Ptr || kind == reflect.Interface ||
			kind == reflect.Func {
			continue
		}

		// Special handling for time.Time struct
		isTimeField := field.Type.PkgPath() == "time" && field.Type.Name() == "Time"

		// Skip other structs that aren't time.Time
		if kind == reflect.Struct && !isTimeField {
			continue
		}

		fields = append(fields, field.Name)

		// Handle time.Time as TIMESTAMP
		if isTimeField {
			types = append(types, "TIMESTAMP")
			defaults = append(defaults, cmp.Or(field.Tag.Get("default"), "NULL"))
		} else {
			switch kind {
			case reflect.String:
				types = append(types, "TEXT")
				defaults = append(defaults, cmp.Or(field.Tag.Get("default"), "''"))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				types = append(types, "INTEGER")
				defaults = append(defaults, cmp.Or(field.Tag.Get("default"), "0"))
			case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
				types = append(types, "REAL")
				defaults = append(defaults, cmp.Or(field.Tag.Get("default"), "0"))
			case reflect.Bool:
				types = append(types, "BOOLEAN")
				defaults = append(defaults, cmp.Or(field.Tag.Get("default"), "FALSE"))
			default:
				types = append(types, "ANY")
				defaults = append(defaults, cmp.Or(field.Tag.Get("default"), "NULL"))
			}
		}
	}

	return
}

func (db *DynamicDB) Reflect(ent Entity) (fields []string, values []any, addrs []any) {
	value := reflect.ValueOf(ent)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	fields, _, _ = db.Fields(ent)
	fields = append([]string{"ID", "CreatedAt", "UpdatedAt"}, fields...)
	for _, field := range fields {
		if !value.FieldByName(field).IsValid() {
			continue
		}
		values = append(values, value.FieldByName(field).Interface())
		addrs = append(addrs, value.FieldByName(field).Addr().Interface())
	}

	return
}

func (db *DynamicDB) qualified(ent Entity, fields []string) (res []string) {
	res = []string{}
	for _, field := range fields {
		res = append(res, fmt.Sprintf("%s.%s", ent.Table(), field))
	}
	return
}

func (db *DynamicDB) entID(ent Entity) (id string) {
	value := reflect.ValueOf(ent)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return
	}

	type_ := value.Type()
	for i := range type_.NumField() {
		field := type_.Field(i)
		if field.Name == "ID" {
			return value.FieldByName(field.Name).String()
		}
	}

	return
}

func (db *DynamicDB) Insert(ent Entity) error {
	fields, values, addrs := db.Reflect(ent)

	places := make([]string, len(fields))
	for i := range fields {
		places[i] = "?"
	}

	return db.Query(fmt.Sprintf(`
		INSERT INTO %[1]s (%[2]s)
		VALUES (%[3]s)
		RETURNING %[2]s
	`, ent.Table(),
		strings.Join(fields, ", "),
		strings.Join(places, ", ")),
		values...).Scan(addrs...)
}

func (db *DynamicDB) Get(id string, ent Entity) error {
	fields, _, addrs := db.Reflect(ent)
	fields = db.qualified(ent, fields)

	places := make([]string, len(fields))
	for i := range fields {
		places[i] = "?"
	}

	return db.Query(fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE ID = ?
	`, strings.Join(fields, ", "), ent.Table()),
		id).Scan(addrs...)
}

func (db *DynamicDB) Update(ent Entity) error {
	var (
		entityID  any
		updatedAt any

		fields, values, addrs = db.Reflect(ent)
	)

	sets := make([]string, len(fields)+1)
	for i, field := range fields {
		switch field {
		case "ID":
			entityID = values[i]
		case "UpdatedAt":
			updatedAt = addrs[i]
		}
		sets[i] = fmt.Sprintf("%s = ?", field)
	}

	sets[len(fields)] = "UpdatedAt = CURRENT_TIMESTAMP"
	return db.Query(fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE ID = ?
		RETURNING UpdatedAt
	`, ent.Table(), strings.Join(sets, ", ")),
		append(values, entityID)...).Scan(updatedAt)
}

func (db *DynamicDB) Delete(ent Entity) error {
	return db.Query(fmt.Sprintf(`
		DELETE FROM %s
		WHERE ID = ?
	`, ent.Table()), db.entID(ent)).Exec()
}

// Index creates an index on the specified table and columns
// Example: db.Index("users", "email") creates idx_users_email
func (db *DynamicDB) Index(table string, columns ...string) error {
	if len(columns) == 0 {
		return nil
	}
	
	// Auto-generate index name from table and columns
	indexName := fmt.Sprintf("idx_%s_%s", table, strings.Join(columns, "_"))
	indexName = strings.ToLower(strings.ReplaceAll(indexName, " ", "_"))
	
	return db.Query(fmt.Sprintf(
		"CREATE INDEX IF NOT EXISTS %s ON %s(%s)",
		indexName, table, strings.Join(columns, ", "),
	)).Exec()
}

// UniqueIndex creates a unique index on the specified table and columns
// Example: db.UniqueIndex("users", "email") creates uniq_users_email
func (db *DynamicDB) UniqueIndex(table string, columns ...string) error {
	if len(columns) == 0 {
		return nil
	}
	
	// Auto-generate index name from table and columns
	indexName := fmt.Sprintf("uniq_%s_%s", table, strings.Join(columns, "_"))
	indexName = strings.ToLower(strings.ReplaceAll(indexName, " ", "_"))
	
	return db.Query(fmt.Sprintf(
		"CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(%s)",
		indexName, table, strings.Join(columns, ", "),
	)).Exec()
}

func Cursor[E Entity](db *DynamicDB, ent E, query string, args ...any) *cursor[E] {
	typeOf := reflect.TypeOf(ent)
	return &cursor[E]{db, typeOf, ent, query, args}
}

type cursor[E Entity] struct {
	db     *DynamicDB
	typeOf reflect.Type
	entity E
	query  string
	args   []any
}

func (c *cursor[E]) Iter(visit func(func(Entity) error) error) error {
	fields, _, _ := c.db.Reflect(c.entity)
	fields = c.db.qualified(c.entity, fields)
	err := c.db.Query(
		fmt.Sprintf(`SELECT %s FROM %s %s`,
			strings.Join(fields, ", "), c.entity.Table(), c.query,
		), c.args...).
		All(func(scan ScanFunc) error {
			return visit(func(ent Entity) error {
				_, _, attrs := c.db.Reflect(ent)
				return scan(attrs...)
			})
		})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (c *cursor[E]) One() (E, error) {
	ent := reflect.New(c.typeOf.Elem()).Interface().(E)
	ent.GetModel().SetDB(c.db)
	fields, _, attrs := c.db.Reflect(ent)
	return ent, c.db.Query(
		fmt.Sprintf(`SELECT %s FROM %s %s`,
			strings.Join(fields, ", "), ent.Table(), c.query,
		), c.args...).
		Scan(attrs...)
}
