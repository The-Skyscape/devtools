package database

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Repository[E Entity] struct {
	DB   *DynamicDB
	Ent  E
	Type reflect.Type
}

func Manage[E Entity](db *DynamicDB, ent E) *Repository[E] {
	db.Register(ent)
	t := reflect.TypeOf(ent)
	db.Repos[ent.Table()] = &Repository[Entity]{db, ent, t}
	return &Repository[E]{db, ent, t}
}

func (r *Repository[E]) Count() (count int) {
	r.DB.Query(`select count(*) from ` + r.Ent.Table()).
		Scan(&count)
	return count
}

func (r *Repository[E]) New() E {
	ent := reflect.New(r.Type.Elem()).Interface().(E)
	ent.GetModel().SetDB(r.DB)
	return ent
}

func (r *Repository[E]) Get(id string) (E, error) {
	ent := r.New()
	return ent, r.DB.Get(id, ent)
}

func (r *Repository[E]) Insert(ent E) (E, error) {
	ent.GetModel().SetDB(r.DB)
	if ent.GetModel().ID == "" {
		ent.GetModel().ID = uuid.NewString()
		ent.GetModel().CreatedAt = time.Now()
		ent.GetModel().UpdatedAt = time.Now()
	}
	return ent, r.DB.Insert(ent)
}

func (r *Repository[E]) Update(ent E) error {
	return r.DB.Update(ent)
}

func (r *Repository[E]) Delete(ent E) error {
	return r.DB.Delete(ent)
}

func (r *Repository[E]) Search(query string, args ...any) ([]E, error) {
	apps := []E{}
	return apps, Cursor(r.DB, r.Ent, query, args...).
		Iter(func(load func(Entity) error) error {
			app := r.New()
			if err := load(app); err != nil {
				return err
			}
			apps = append(apps, app)
			return nil
		})
}
