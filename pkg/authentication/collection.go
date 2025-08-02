package authentication

import (
	"errors"
	"fmt"

	"github.com/The-Skyscape/devtools/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func Manage(db *database.DynamicDB) *Collection {
	repo := &Collection{
		db:       db,
		Users:    database.Manage(db, new(User)),
		Sessions: database.Manage(db, new(Session)),
	}

	db.Query(`
		CREATE UNIQUE INDEX IF NOT EXISTS unique_handle ON users (handle);
		CREATE UNIQUE INDEX IF NOT EXISTS unique_email ON users (email);
	`).Exec()

	return repo
}

type Collection struct {
	db       *database.DynamicDB
	Users    *database.Collection[*User]
	Sessions *database.Collection[*Session]
}

func (c *Collection) GetUser(ident string) (*User, error) {
	return database.Cursor(c.db, new(User), `

		WHERE ID = $1 OR Email = $1 OR Handle = $1
	
	`, ident).One()
}

func (c *Collection) Signup(name, email, handle, password string, isAdmin bool) (*User, error) {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return c.Users.Insert(&User{
		Avatar:   fmt.Sprintf("https://robohash.org/%s?set=set4", email),
		Name:     name,
		Email:    email,
		Handle:   handle,
		PassHash: passhash,
		IsAdmin:  isAdmin,
	})
}

func (c *Collection) Signin(ident string, password string) (user *User, err error) {
	if user, err = c.GetUser(ident); err != nil {
		return nil, errors.New("user not found")
	}

	if !user.VerifyPassword(password) {
		return nil, errors.New("user not found")
	}

	return user, nil
}
