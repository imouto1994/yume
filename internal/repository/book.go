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
	UpdateModifiedTime(context.Context, sqlite.DBOps, string, string) error
	UpdatePreview(context.Context, sqlite.DBOps, string, *string, *string) error
	UpdateFormat(context.Context, sqlite.DBOps, string, string) error
	UpdatePageCount(context.Context, sqlite.DBOps, string, int) error
	DeleteAllByTitleID(context.Context, sqlite.DBOps, string) error
	DeleteAllByLibraryID(context.Context, sqlite.DBOps, string) error
	DeleteByID(context.Context, sqlite.DBOps, string) error
}

type repositoryBook struct {
}

func NewRepositoryBook() RepositoryBook {
	return &repositoryBook{}
}

func (r *repositoryBook) Insert(ctx context.Context, db sqlite.DBOps, book *model.Book) error {
	query := "INSERT INTO BOOK (NAME, URL, CREATED_AT, UPDATED_AT, PREVIEW_URL, PREVIEW_UPDATED_AT, PAGE_COUNT, FORMAT, TITLE_ID, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := db.ExecContext(ctx, query, book.Name, book.URL, book.CreatedAt, book.UpdatedAt, book.PreviewURL, book.PreviewUpdatedAt, book.PageCount, book.Format, book.TitleID, book.LibraryID)
	if err != nil {
		return fmt.Errorf("rBook - failed to add new row to table BOOK: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("rBook - failed to get the ID of inserted row: %w", err)
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
		return nil, fmt.Errorf("rBook- %w: no matched rows with specific ID from table BOOK", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("rBook - failed to find row with specific ID from table BOOK: %w", err)
	}

	return &book, nil
}

func (r *repositoryBook) FindAllByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) ([]*model.Book, error) {
	query := "SELECT * FROM BOOK " +
		"WHERE TITLE_ID = ? " +
		"ORDER BY NAME ASC"

	books := []*model.Book{}

	err := dbOps.SelectContext(ctx, &books, query, titleID)
	if err != nil {
		return nil, fmt.Errorf("rBook - failed to find rows with specific TITLE_ID from table BOOK: %w", err)
	}

	return books, nil
}

func (r *repositoryBook) UpdateModifiedTime(ctx context.Context, dbOps sqlite.DBOps, bookID string, modTime string) error {
	query := "UPDATE BOOK " +
		"SET UPDATED_AT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, modTime, bookID)
	if err != nil {
		return fmt.Errorf("rBook - failed to update UPDATED_AT field for row with given ID from table BOOK: %w", err)
	}

	return nil
}

func (r *repositoryBook) UpdatePreview(ctx context.Context, dbOps sqlite.DBOps, bookID string, previewURL *string, previewUpdatedAt *string) error {
	query := "UPDATE BOOK " +
		"SET PREVIEW_URL = ?, PREVIEW_UPDATED_AT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, previewURL, previewUpdatedAt, bookID)
	if err != nil {
		return fmt.Errorf("rBook - failed to update PREVIEW_URL & PREVIEW_UPDATED_AT fields for row with given ID from table BOOK: %w", err)
	}

	return nil
}

func (r *repositoryBook) UpdateFormat(ctx context.Context, dbOps sqlite.DBOps, bookID string, format string) error {
	query := "UPDATE BOOK " +
		"SET FORMAT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, format, bookID)
	if err != nil {
		return fmt.Errorf("rBook - failed to update FORMAT field for row with given ID from table BOOK: %w", err)
	}

	return nil
}

func (r *repositoryBook) UpdatePageCount(ctx context.Context, dbOps sqlite.DBOps, bookID string, pageCount int) error {
	query := "UPDATE BOOK " +
		"SET PAGE_COUNT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, pageCount, bookID)
	if err != nil {
		return fmt.Errorf("rBook - failed to update PAGE_COUNT field for row with given ID from table BOOK: %w", err)
	}

	return nil
}

func (r *repositoryBook) DeleteAllByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) error {
	query := "DELETE FROM BOOK " +
		"WHERE TITLE_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, titleID)
	if err != nil {
		return fmt.Errorf("rBook - failed to delete rows with given BOOK_ID from table BOOK: %w", err)
	}

	return nil
}

func (r *repositoryBook) DeleteAllByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	query := "DELETE FROM BOOK " +
		"WHERE LIBRARY_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, libraryID)
	if err != nil {
		return fmt.Errorf("rBook - failed to delete rows with given LIBRARY_ID from table BOOK: %w", err)
	}

	return nil
}

func (r *repositoryBook) DeleteByID(ctx context.Context, dbOps sqlite.DBOps, bookID string) error {
	query := "DELETE FROM BOOK " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("rBook - failed to delete row with given ID from table BOOK: %w", err)
	}

	return nil
}
