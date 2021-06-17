package repository

import (
	"fmt"

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
	addQuery := `INSERT INTO library (name, root) VALUES (?, ?)`

	result, err := r.db.Exec(addQuery, library.Name, library.Root)
	if err != nil {
		return fmt.Errorf("Failed to add new row to table LIBRARY: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("Failed to get the ID of inserted row: %w", err)
	}

	library.ID = rowID

	return nil
}

func (r *LibraryRepository) GetLibraries() ([]*model.Library, error) {
	return nil, nil
}
