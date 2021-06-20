package model

type Title struct {
	ID        int64  `json:"id" db:"ID"`
	Name      string `json:"name" db:"NAME"`
	URL       string `json:"url" db:"URL"`
	LibraryID string `json:"library_id" db:"LIBRARY_ID"`
	CreatedAt string `json:"created_at" db:"CREATED_AT"`
	UpdatedAt string `json:"updated_at" db:"UPDATED_AT"`
}
