package repository

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type RepositoryPreview interface {
	Insert(context.Context, sqlite.DBOps, *model.Preview) error
	FindAllByBookID(context.Context, sqlite.DBOps, string) ([]*model.Preview, error)
	DeleteAllByBookID(context.Context, sqlite.DBOps, string) error
	DeleteAllByTitleID(context.Context, sqlite.DBOps, string) error
	DeleteAllByLibraryID(context.Context, sqlite.DBOps, string) error
}

type repositoryPreview struct {
}

func NewRepositoryPreview() RepositoryPreview {
	return &repositoryPreview{}
}

func (r *repositoryPreview) Insert(ctx context.Context, dbOps sqlite.DBOps, preview *model.Preview) error {
	query := "INSERT INTO PREVIEW (FILE_INDEX, NUMBER, BOOK_ID, TITLE_ID, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?)"

	_, err := dbOps.ExecContext(ctx, query, preview.Index, preview.Number, preview.BookID, preview.TitleID, preview.LibraryID)
	if err != nil {
		return fmt.Errorf("rPreview - failed to add new row to table PREVIEW: %w", err)
	}

	return nil
}

func (r *repositoryPreview) FindAllByBookID(ctx context.Context, dbOps sqlite.DBOps, bookID string) ([]*model.Preview, error) {
	query := "SELECT * FROM PREVIEW " +
		"WHERE BOOK_ID = ? " +
		"ORDER BY NUMBER ASC"

	previews := []*model.Preview{}

	err := dbOps.SelectContext(ctx, &previews, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("rPreview -failed to find rows with specific BOOK_ID from table PREVIEW: %w", err)
	}

	return previews, nil
}

func (r *repositoryPreview) DeleteAllByBookID(ctx context.Context, dbOps sqlite.DBOps, bookID string) error {
	query := "DELETE FROM PREVIEW " +
		"WHERE BOOK_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("rPreview -failed to delete rows with given BOOK_ID from table PREVIEW: %w", err)
	}

	return nil
}

func (r *repositoryPreview) DeleteAllByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) error {
	query := "DELETE FROM PREVIEW " +
		"WHERE TITLE_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, titleID)
	if err != nil {
		return fmt.Errorf("rPreview - failed to delete rows with given TITLE_ID from table PREVIEW: %w", err)
	}

	return nil
}

func (r *repositoryPreview) DeleteAllByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	query := "DELETE FROM PREVIEW " +
		"WHERE LIBRARY_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, libraryID)
	if err != nil {
		return fmt.Errorf("rPreview -failed to delete rows with given LIBRARY_ID from table PREVIEW: %w", err)
	}

	return nil
}
