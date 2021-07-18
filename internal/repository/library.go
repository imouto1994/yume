package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type RepositoryLibrary interface {
	Insert(context.Context, sqlite.DBOps, *model.Library) error
	FindAll(context.Context, sqlite.DBOps) ([]*model.Library, error)
	FindByID(context.Context, sqlite.DBOps, string) (*model.Library, error)
	DeleteByID(context.Context, sqlite.DBOps, string) error
}

type repositoryLibrary struct {
}

func NewRepositoryLibrary() RepositoryLibrary {
	return &repositoryLibrary{}
}

func (r *repositoryLibrary) Insert(ctx context.Context, dbOps sqlite.DBOps, library *model.Library) error {
	query := "INSERT INTO LIBRARY (NAME, ROOT) " +
		"VALUES (?, ?)"

	result, err := dbOps.ExecContext(ctx, query, library.Name, library.Root)
	if err != nil {
		return fmt.Errorf("rLibrary - failed to add new row to table LIBRARY: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("rLibrary - failed to get the ID of inserted row: %w", err)
	}

	library.ID = rowID

	return nil
}

func (r *repositoryLibrary) FindAll(ctx context.Context, dbOps sqlite.DBOps) ([]*model.Library, error) {
	query := "SELECT * FROM LIBRARY"
	libraries := []*model.Library{}

	err := dbOps.SelectContext(ctx, &libraries, query)
	if err != nil {
		return nil, fmt.Errorf("rLibrary - failed to find all rows from table LIBRARY: %w", err)
	}

	return libraries, nil
}

func (r *repositoryLibrary) FindByID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) (*model.Library, error) {
	query := "SELECT * FROM LIBRARY " +
		"WHERE ID = ?"

	library := model.Library{}

	err := dbOps.GetContext(ctx, &library, query, libraryID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rLibrary - %w: no matched rows with specific ID from table LIBRARY", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("rLibrary - failed to find row with specific ID from table LIBRARY: %w", err)
	}

	return &library, nil
}

func (r *repositoryLibrary) DeleteByID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	query := "DELETE FROM LIBRARY " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, libraryID)
	if err != nil {
		return fmt.Errorf("rLibrary - failed to delete row with given ID from table LIBRARY: %w", err)
	}

	return nil
}
