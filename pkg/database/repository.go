package database

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Collection[E Entity] struct {
	DB   *DynamicDB
	Ent  E
	Type reflect.Type
}

func Manage[E Entity](db *DynamicDB, ent E) *Collection[E] {
	db.Register(ent)
	t := reflect.TypeOf(ent)
	db.Repos[ent.Table()] = &Collection[Entity]{db, ent, t}
	return &Collection[E]{db, ent, t}
}

func (c *Collection[E]) Count() (count int) {
	c.DB.Query(`select count(*) from ` + c.Ent.Table()).
		Scan(&count)
	return count
}

func (c *Collection[E]) New() E {
	ent := reflect.New(c.Type.Elem()).Interface().(E)
	ent.GetModel().SetDB(c.DB)
	return ent
}

func (c *Collection[E]) Get(id string) (E, error) {
	ent := c.New()
	return ent, c.DB.Get(id, ent)
}

func (c *Collection[E]) Insert(ent E) (E, error) {
	ent.GetModel().SetDB(c.DB)
	if ent.GetModel().ID == "" {
		ent.GetModel().ID = uuid.NewString()
		ent.GetModel().CreatedAt = time.Now()
		ent.GetModel().UpdatedAt = time.Now()
	}
	return ent, c.DB.Insert(ent)
}

func (c *Collection[E]) Update(ent E) error {
	return c.DB.Update(ent)
}

func (c *Collection[E]) Delete(ent E) error {
	return c.DB.Delete(ent)
}

func (c *Collection[E]) Search(query string, args ...any) ([]E, error) {
	apps := []E{}
	return apps, Cursor(c.DB, c.Ent, query, args...).
		Iter(func(load func(Entity) error) error {
			app := c.New()
			if err := load(app); err != nil {
				return err
			}
			apps = append(apps, app)
			return nil
		})
}

func (c *Collection[E]) Find(query string, args ...any) (E, error) {
	app := c.New()
	return app, Cursor(c.DB, c.Ent, query, args...).
		Iter(func(load func(Entity) error) error {
			if err := load(app); err != nil {
				return err
			}
			return ErrIterStop
		})
}
