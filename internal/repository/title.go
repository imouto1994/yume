package repository

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/jmoiron/sqlx"
)

type TitleRepository struct {
}

func NewTitleRepository(db *sqlx.DB) *TitleRepository {
	return &TitleRepository{}
}

func (r *TitleRepository) Insert(ctx context.Context, db sqlite.DBOps, title *model.Title) error {
	query := "INSERT INTO TITLE (NAME, URL, CREATED_AT, UPDATED_AT, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?)"

	result, err := db.ExecContext(ctx, query, title.Name, title.URL, title.CreatedAt, title.UpdatedAt, title.LibraryID)
	if err != nil {
		return fmt.Errorf("failed to add new row to table TITLE: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get the ID of inserted row: %w", err)
	}

	title.ID = rowID

	return nil
}
