package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type RepositoryPage interface {
	Insert(context.Context, sqlite.DBOps, *model.Page) error
	FindAllByBookID(context.Context, sqlite.DBOps, string) ([]*model.Page, error)
}

type repositoryPage struct {
}

func NewRepositoryPage() RepositoryPage {
	return &repositoryPage{}
}

func (r *repositoryPage) Insert(ctx context.Context, db sqlite.DBOps, page *model.Page) error {
	query := "INSERT INTO PAGE (FILE_INDEX, NUMBER, WIDTH, HEIGHT, BOOK_ID) " +
		"VALUES (?, ?, ?, ?, ?)"

	_, err := db.ExecContext(ctx, query, page.Index, page.Number, page.Width, page.Height, page.BookID)
	if err != nil {
		return fmt.Errorf("failed to add new row to table PAGE: %w", err)
	}

	return nil
}

func (r *repositoryPage) FindAllByBookID(ctx context.Context, dbOps sqlite.DBOps, bookID string) ([]*model.Page, error) {
	query := "SELECT * FROM PAGE " +
		"WHERE BOOK_ID = ? " +
		"ORDER BY NUMBER ASC"

	pages := []*model.Page{}

	err := dbOps.SelectContext(ctx, &pages, query, bookID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: no matched rows with specific BOOK_ID from table PAGE", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find rows with specific BOOK_ID from table PAGE: %w", err)
	}

	return pages, nil
}
