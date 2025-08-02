package authentication

import (
	"github.com/The-Skyscape/devtools/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func (*User) Table() string { return "users" }

type User struct {
	*Collection

	database.Model
	Avatar   string
	Name     string
	Email    string
	Handle   string
	IsAdmin  bool
	PassHash []byte
}

func (user *User) SetupPassword(password string) (err error) {
	user.PassHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return user.Users.Update(user)
}

func (user *User) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)) == nil
}
