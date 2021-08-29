package model

type Preview struct {
	Index     int   `json:"index" db:"FILE_INDEX"`
	Number    int   `json:"number" db:"NUMBER"`
	BookID    int64 `json:"book_id" db:"BOOK_ID"`
	TitleID   int64 `json:"title_id" db:"TITLE_ID"`
	LibraryID int64 `json:"library_id" db:"LIBRARY_ID"`
}
