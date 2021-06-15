package model

import "time"

type Book struct {
	ID string
	Name string
	URL string
	TitleID string
	LibraryID string
	PageCount int64
	CreatedDate time.Time
	LastModifiedDate time.Time
}
