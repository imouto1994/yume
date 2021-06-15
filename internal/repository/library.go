package repository

import (
	"github.com/imouto1994/yume/internal/model"
	"github.com/jmoiron/sqlx"
)

type LibraryRepository struct {
	db *sqlx.DB
}

func NewLibraryRepository(db *sqlx.DB) *LibraryRepository {
	return &LibraryRepository{
		db: db,
	}
}

func (r *LibraryRepository) Add(library *model.Library) error {
	return nil
}

func (r *LibraryRepository) GetLibraries() ([]*model.Library, error) {
	return nil, nil
}
