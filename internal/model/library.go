package model

type Library struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Root string `json:"root" db:"root"`
}
