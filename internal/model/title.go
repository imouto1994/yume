package model

import "time"

type Title struct {
	ID               string
	Name             string
	URL              string
	LibraryID        string
	CreatedDate      time.Time
	LastModifiedDate time.Time
}
