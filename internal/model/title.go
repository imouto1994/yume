package model

type Title struct {
	ID          int64  `json:"id" db:"ID"`
	Name        string `json:"name" db:"NAME"`
	URL         string `json:"url" db:"URL"`
	CreatedAt   string `json:"created_at" db:"CREATED_AT"`
	UpdatedAt   string `json:"updated_at" db:"UPDATED_AT"`
	CoverWidth  int    `json:"cover_width" db:"COVER_WIDTH"`
	CoverHeight int    `json:"cover_height" db:"COVER_HEIGHT"`
	BookCount   int    `json:"book_count" db:"BOOK_COUNT"`
	Uncensored  int    `json:"uncensored" db:"UNCENSORED"`
	Langs       string `json:"langs" db:"LANGS"`
	LibraryID   int64  `json:"library_id" db:"LIBRARY_ID"`
}

type TitleQuery struct {
	LibraryIDs []string
	Page       int
	Size       int
	Sort       string
	Search     string
}
