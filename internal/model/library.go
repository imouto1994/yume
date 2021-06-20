package model

type Library struct {
	ID   int64  `json:"id" db:"ID"`
	Name string `json:"name" db:"NAME"`
	Root string `json:"root" db:"ROOT"`
}
