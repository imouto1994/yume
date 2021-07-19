package repository

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type RepositoryPage interface {
	Insert(context.Context, sqlite.DBOps, *model.Page) error
	FindAllByBookID(context.Context, sqlite.DBOps, string) ([]*model.Page, error)
	DeleteAllByBookID(context.Context, sqlite.DBOps, string) error
	DeleteAllByTitleID(context.Context, sqlite.DBOps, string) error
	DeleteAllByLibraryID(context.Context, sqlite.DBOps, string) error
}

type repositoryPage struct {
}

func NewRepositoryPage() RepositoryPage {
	return &repositoryPage{}
}

func (r *repositoryPage) Insert(ctx context.Context, dbOps sqlite.DBOps, page *model.Page) error {
	query := "INSERT INTO PAGE (FILE_INDEX, NUMBER, WIDTH, HEIGHT, BOOK_ID, TITLE_ID, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?)"

	_, err := dbOps.ExecContext(ctx, query, page.Index, page.Number, page.Width, page.Height, page.BookID, page.TitleID, page.LibraryID)
	if err != nil {
		return fmt.Errorf("rPage - failed to add new row to table PAGE: %w", err)
	}

	return nil
}

func (r *repositoryPage) FindAllByBookID(ctx context.Context, dbOps sqlite.DBOps, bookID string) ([]*model.Page, error) {
	query := "SELECT * FROM PAGE " +
		"WHERE BOOK_ID = ? " +
		"ORDER BY NUMBER ASC"

	pages := []*model.Page{}

	err := dbOps.SelectContext(ctx, &pages, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("rPage -failed to find rows with specific BOOK_ID from table PAGE: %w", err)
	}

	return pages, nil
}

func (r *repositoryPage) DeleteAllByBookID(ctx context.Context, dbOps sqlite.DBOps, bookID string) error {
	query := "DELETE FROM PAGE " +
		"WHERE BOOK_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("rPage -failed to delete rows with given BOOK_ID from table PAGE: %w", err)
	}

	return nil
}

func (r *repositoryPage) DeleteAllByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) error {
	query := "DELETE FROM PAGE " +
		"WHERE TITLE_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, titleID)
	if err != nil {
		return fmt.Errorf("rPage - failed to delete rows with given TITLE_ID from table PAGE: %w", err)
	}

	return nil
}

func (r *repositoryPage) DeleteAllByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	query := "DELETE FROM PAGE " +
		"WHERE LIBRARY_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, libraryID)
	if err != nil {
		return fmt.Errorf("rPage -failed to delete rows with given LIBRARY_ID from table PAGE: %w", err)
	}

	return nil
}
