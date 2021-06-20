package repository

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type LibraryRepository struct {
}

func NewLibraryRepository() *LibraryRepository {
	return &LibraryRepository{}
}

func (r *LibraryRepository) Insert(ctx context.Context, db sqlite.DBOps, library *model.Library) error {
	query := "INSERT INTO LIBRARY (NAME, ROOT) " +
		"VALUES (?, ?)"

	result, err := db.ExecContext(ctx, query, library.Name, library.Root)
	if err != nil {
		return fmt.Errorf("failed to add new row to table LIBRARY: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get the ID of inserted row: %w", err)
	}

	library.ID = rowID

	return nil
}

func (r *LibraryRepository) FindAll(ctx context.Context, db sqlite.DBOps) ([]*model.Library, error) {
	query := "SELECT * FROM LIBRARY"
	libraries := []*model.Library{}

	err := db.SelectContext(ctx, &libraries, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find all rows from table LIBRARY: %w", err)
	}

	return libraries, nil
}

func (r *LibraryRepository) Delete(ctx context.Context, db sqlite.DBOps, libraryID string) error {
	query := "DELETE FROM library " +
		"WHERE ID = ?"

	_, err := db.ExecContext(ctx, query, libraryID)
	if err != nil {
		return fmt.Errorf("failed to delete row from table LIBRARY: %w", err)
	}

	return nil
}
