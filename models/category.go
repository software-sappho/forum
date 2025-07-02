package models

type Category struct {
	ID          int    `db:"id" json:"ID"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}
