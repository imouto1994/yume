package model

import (
	"time"
)

type Library struct {
	ID string
	Name string
	Root string
	CreatedDate time.Time
	LastModifiedDate time.Time
}
