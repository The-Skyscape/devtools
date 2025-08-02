package authentication

import (
	"errors"
	"fmt"
	"github.com/The-Skyscape/devtools/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func Manage(db *database.DynamicDB) *Repository {
	repo := &Repository{
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

type Repository struct {
	db       *database.DynamicDB
	Users    *database.Repository[*User]
	Sessions *database.Repository[*Session]
}

func (r *Repository) GetUser(ident string) (*User, error) {
	return database.Cursor(r.db, new(User), `

		WHERE ID = $1 OR Email = $1 OR Handle = $1
	
	`, ident).One()
}

func (r *Repository) Signup(name, email, handle, password string, isAdmin bool) (*User, error) {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return r.Users.Insert(&User{
		Avatar:   fmt.Sprintf("https://robohash.org/%s?set=set4", email),
		Name:     name,
		Email:    email,
		Handle:   handle,
		PassHash: passhash,
		IsAdmin:  isAdmin,
	})
}

func (r *Repository) Signin(ident string, password string) (user *User, err error) {
	if user, err = r.GetUser(ident); err != nil {
		return nil, errors.New("user not found")
	}

	if !user.VerifyPassword(password) {
		return nil, errors.New("user not found")
	}

	return user, nil
}
