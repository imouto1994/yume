package repository

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/jmoiron/sqlx"
)

type BookRepository struct {
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{}
}

func (r *BookRepository) Insert(ctx context.Context, db sqlite.DBOps, book *model.Book) error {
	query := "INSERT INTO BOOK (NAME, URL, CREATED_AT, UPDATED_AT, PAGE_COUNT, TITLE_ID, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?)"

	result, err := db.ExecContext(ctx, query, book.Name, book.URL, book.CreatedAt, book.UpdatedAt, book.PageCount, book.TitleID, book.LibraryID)
	if err != nil {
		return fmt.Errorf("failed to add new row to table BOOK: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get the ID of inserted row: %w", err)
	}

	book.ID = rowID

	return nil
}
