package model

type Title struct {
	ID          int64  `json:"id" db:"ID"`
	Name        string `json:"name" db:"NAME"`
	URL         string `json:"url" db:"URL"`
	LibraryID   int64  `json:"library_id" db:"LIBRARY_ID"`
	CreatedAt   string `json:"created_at" db:"CREATED_AT"`
	UpdatedAt   string `json:"updated_at" db:"UPDATED_AT"`
	CoverWidth  int    `json:"cover_width" db:"COVER_WIDTH"`
	CoverHeight int    `json:"cover_height" db:"COVER_HEIGHT"`
}
