package model

type Page struct {
	Index  int   `json:"index" db:"FILE_INDEX"`
	Number int   `json:"number" db:"NUMBER"`
	BookID int64 `json:"book_id" db:"BOOK_ID"`
	Width  int   `json:"width" db:"WIDTH"`
	Height int   `json:"height" db:"HEIGHT"`
}
