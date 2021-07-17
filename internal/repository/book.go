package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type RepositoryBook interface {
	Insert(context.Context, sqlite.DBOps, *model.Book) error
	FindByID(context.Context, sqlite.DBOps, string) (*model.Book, error)
	FindAllByTitleID(context.Context, sqlite.DBOps, string) ([]*model.Book, error)
}

type repositoryBook struct {
}

func NewRepositoryBook() RepositoryBook {
	return &repositoryBook{}
}

func (r *repositoryBook) Insert(ctx context.Context, db sqlite.DBOps, book *model.Book) error {
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

func (r *repositoryBook) FindByID(ctx context.Context, dbOps sqlite.DBOps, bookID string) (*model.Book, error) {
	query := "SELECT * FROM BOOK " +
		"WHERE ID = ?"

	book := model.Book{}

	err := dbOps.GetContext(ctx, &book, query, bookID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: no matched rows with specific ID from table BOOK", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find row with specific ID from table BOOK: %w", err)
	}

	return &book, nil
}

func (r *repositoryBook) FindAllByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) ([]*model.Book, error) {
	query := "SELECT * FROM BOOK " +
		"WHERE TITLE_ID = ?"

	books := []*model.Book{}

	err := dbOps.GetContext(ctx, &books, query, titleID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: no matched rows with specific TITLE_ID from table BOOK", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find rows with specific TITLE_ID from table BOOK: %w", err)
	}

	return books, nil
}
