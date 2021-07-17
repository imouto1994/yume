package model

type Book struct {
	ID        int64  `json:"id" db:"ID"`
	Name      string `json:"name" db:"NAME"`
	URL       string `json:"url" db:"URL"`
	TitleID   int64  `json:"title_id" db:"TITLE_ID"`
	LibraryID int64  `json:"library_id" db:"LIBRARY_ID"`
	CreatedAt string `json:"created_at" db:"CREATED_AT"`
	UpdatedAt string `json:"updated_at" db:"UPDATED_AT"`
	PageCount int    `json:"page_count" db:"PAGE_COUNT"`
}
