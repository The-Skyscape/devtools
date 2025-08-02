package models

import "github.com/The-Skyscape/devtools/pkg/application"

// Table is the name of the table where ducks are stored
func (*Duck) Table() string { return "ducks" }

// Duck is the model for storing ducks
type Duck struct {
	application.Model
	Name  string
	Breed string
}

// GetDuckByID returns a duck by its ID
func GetDuckByID(id string) (*Duck, error) {
	return Ducks.Get(id)
}

// GetDuckByBreed returns a list of ducks
func DucksByBreed(breed string) ([]*Duck, error) {
	return Ducks.Search("breed = ?", breed)
}

// GetDuckByName returns a single named duck
func DucksByName(name string) (*Duck, error) {
	return Ducks.Find("name = ?", name)
}

// Quack is a public method that can be called in views
func (d *Duck) Quack() string {
	return "Quack"
}
