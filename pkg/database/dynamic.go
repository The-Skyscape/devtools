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
	Repos map[string]*Repository[Entity]
}

type Entity interface {
	Table() string
	GetModel() *Model
}

func Dynamic(engine Database, opts ...DynamicDBOption) *DynamicDB {
	db := DynamicDB{engine, []Entity{}, map[string]*Repository[Entity]{}}
	for _, opt := range opts {
		opt(&db)
	}
	return &db
}

type DynamicDBOption func(*DynamicDB)

func WithModel(ent Entity) DynamicDBOption {
	return func(db *DynamicDB) {
		if err := db.Register(ent); err != nil {
			log.Fatalf("failed to register model %v", err)
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
		if field.Anonymous || kind == reflect.Ptr || kind == reflect.Interface ||
			kind == reflect.Func || kind == reflect.Struct {
			continue
		}

		fields = append(fields, field.Name)
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
		DELETE %s
		WHERE ID = ?
	`, ent.Table()), db.entID(ent)).Exec()
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
