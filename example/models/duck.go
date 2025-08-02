package models


type Duck struct {
	database.Model
	Name string
	Breed string
}