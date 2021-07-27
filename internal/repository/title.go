package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/jmoiron/sqlx"
)

type RepositoryTitle interface {
	Insert(context.Context, sqlite.DBOps, *model.Title) error
	Find(context.Context, sqlite.DBOps, *model.TitleQuery) ([]*model.Title, error)
	GetTotalFindResults(context.Context, sqlite.DBOps, *model.TitleQuery) (int, error)
	FindAllByLibraryID(context.Context, sqlite.DBOps, string) ([]*model.Title, error)
	FindByID(context.Context, sqlite.DBOps, string) (*model.Title, error)
	UpdateModifiedTime(context.Context, sqlite.DBOps, string, string) error
	UpdateCoverDimension(context.Context, sqlite.DBOps, string, int, int) error
	UpdateBookCount(context.Context, sqlite.DBOps, string, int) error
	UpdateUncensored(context.Context, sqlite.DBOps, string, int) error
	UpdateLangs(context.Context, sqlite.DBOps, string, string) error
	DeleteAllByLibraryID(context.Context, sqlite.DBOps, string) error
	DeleteByID(context.Context, sqlite.DBOps, string) error
}

type repositoryTitle struct {
}

func NewRepositoryTitle() RepositoryTitle {
	return &repositoryTitle{}
}

func (r *repositoryTitle) Insert(ctx context.Context, db sqlite.DBOps, title *model.Title) error {
	query := "INSERT INTO TITLE (NAME, URL, CREATED_AT, UPDATED_AT, COVER_WIDTH, COVER_HEIGHT, BOOK_COUNT, UNCENSORED, LANGS, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := db.ExecContext(ctx, query, title.Name, title.URL, title.CreatedAt, title.UpdatedAt, title.CoverWidth, title.CoverHeight, title.BookCount, title.Uncensored, title.Langs, title.LibraryID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to add new row to table TITLE: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("rTitle - failed to get the ID of inserted row: %w", err)
	}

	title.ID = rowID

	return nil
}

func (r *repositoryTitle) Find(ctx context.Context, dbOps sqlite.DBOps, titleQuery *model.TitleQuery) ([]*model.Title, error) {
	orderBy := "CREATED_AT DESC"
	if titleQuery.Sort == "name" {
		orderBy = "NAME COLLATE NOCASE ASC"
	} else if titleQuery.Sort == "updated_at" {
		orderBy = "UPDATED_AT DESC"
	}

	// Build SQL query
	var query string
	var args []interface{}
	var err error
	if len(titleQuery.LibraryIDs) > 0 && titleQuery.Search != "" {
		queryString := "SELECT * FROM TITLE " +
			"WHERE LIBRARY_ID IN (?) AND NAME LIKE ? " +
			fmt.Sprintf("ORDER BY %s ", orderBy) +
			"LIMIT ? " +
			"OFFSET ?"
		query, args, err = sqlx.In(queryString, titleQuery.LibraryIDs, "%"+titleQuery.Search+"%", titleQuery.Size, titleQuery.Page*titleQuery.Size)
		if err != nil {
			return nil, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	} else if len(titleQuery.LibraryIDs) > 0 {
		queryString := "SELECT * FROM TITLE " +
			"WHERE LIBRARY_ID IN (?) " +
			fmt.Sprintf("ORDER BY %s ", orderBy) +
			"LIMIT ? " +
			"OFFSET ?"
		query, args, err = sqlx.In(queryString, titleQuery.LibraryIDs, titleQuery.Size, titleQuery.Page*titleQuery.Size)
		if err != nil {
			return nil, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	} else if titleQuery.Search != "" {
		queryString := "SELECT * FROM TITLE " +
			"WHERE NAME LIKE ? " +
			fmt.Sprintf("ORDER BY %s ", orderBy) +
			"LIMIT ? " +
			"OFFSET ?"
		query, args, err = sqlx.In(queryString, "%"+titleQuery.Search+"%", titleQuery.Size, titleQuery.Page*titleQuery.Size)
		if err != nil {
			return nil, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	} else {
		queryString := "SELECT * FROM TITLE " +
			fmt.Sprintf("ORDER BY %s ", orderBy) +
			"LIMIT ? " +
			"OFFSET ?"
		query, args, err = sqlx.In(queryString, titleQuery.Size, titleQuery.Page*titleQuery.Size)
		if err != nil {
			return nil, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	}

	// Execute SQL Query
	titles := []*model.Title{}
	err = dbOps.SelectContext(ctx, &titles, query, args...)
	if err != nil {
		return nil, fmt.Errorf("rTitle - failed to find rows with specific query from table TITLE: %w", err)
	}

	return titles, nil
}

func (r *repositoryTitle) GetTotalFindResults(ctx context.Context, dbOps sqlite.DBOps, titleQuery *model.TitleQuery) (int, error) {
	// Build SQL query
	var query string
	var args []interface{}
	var err error
	if len(titleQuery.LibraryIDs) > 0 && titleQuery.Search != "" {
		queryString := "SELECT COUNT(*) FROM TITLE " +
			"WHERE LIBRARY_ID IN (?) AND NAME LIKE ?"
		query, args, err = sqlx.In(queryString, titleQuery.LibraryIDs, "%"+titleQuery.Search+"%")
		if err != nil {
			return 0, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	} else if len(titleQuery.LibraryIDs) > 0 {
		queryString := "SELECT COUNT(*) FROM TITLE " +
			"WHERE LIBRARY_ID IN (?)"
		query, args, err = sqlx.In(queryString, titleQuery.LibraryIDs)
		if err != nil {
			return 0, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	} else if titleQuery.Search != "" {
		queryString := "SELECT COUNT(*) FROM TITLE " +
			"WHERE NAME LIKE ?"
		query, args, err = sqlx.In(queryString, "%"+titleQuery.Search+"%")
		if err != nil {
			return 0, fmt.Errorf("rTitle - failed to bind variables for SQL query: %w", err)
		}
		query = dbOps.Rebind(query)
	} else {
		query = "SELECT COUNT(*) FROM TITLE"
	}

	// Execute SQL Query
	var count int
	err = dbOps.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("rTitle - failed to count rows with specific query from table TITLE: %w", err)
	}

	return count, nil
}

func (r *repositoryTitle) FindAllByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) ([]*model.Title, error) {
	query := "SELECT * FROM TITLE " +
		"WHERE LIBRARY_ID = ?"

	titles := []*model.Title{}

	err := dbOps.SelectContext(ctx, &titles, query, libraryID)
	if err != nil {
		return nil, fmt.Errorf("rTitle - failed to find rows with specific LIBRARY_ID from table TITLE: %w", err)
	}

	return titles, nil
}

func (r *repositoryTitle) FindByID(ctx context.Context, dbOps sqlite.DBOps, titleID string) (*model.Title, error) {
	query := "SELECT * FROM TITLE " +
		"WHERE ID = ?"

	title := model.Title{}

	err := dbOps.GetContext(ctx, &title, query, titleID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rTitle - %w: no matched rows with specific ID from table TITLE", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("rTitle - failed to find row with specific ID from table TITLE: %w", err)
	}

	return &title, nil
}

func (r *repositoryTitle) UpdateModifiedTime(ctx context.Context, dbOps sqlite.DBOps, titleID string, modTime string) error {
	query := "UPDATE TITLE " +
		"SET UPDATED_AT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, modTime, titleID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to update UPDATED_AT field for row with given ID from table TITLE: %w", err)
	}

	return nil
}

func (r *repositoryTitle) UpdateCoverDimension(ctx context.Context, dbOps sqlite.DBOps, titleID string, coverWidth int, coverHeight int) error {
	query := "UPDATE TITLE " +
		"SET COVER_WIDTH = ?, COVER_HEIGHT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, coverWidth, coverHeight, titleID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to update COVER_WIDTH & COVER_HEIGHT fields for row with given ID from table TITLE: %w", err)
	}

	return nil
}

func (r *repositoryTitle) UpdateBookCount(ctx context.Context, dbOps sqlite.DBOps, titleID string, count int) error {
	query := "UPDATE TITLE " +
		"SET BOOK_COUNT = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, count, titleID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to update BOOK_COUNT field for row with given ID from table TITLE: %w", err)
	}

	return nil
}

func (r *repositoryTitle) UpdateUncensored(ctx context.Context, dbOps sqlite.DBOps, titleID string, uncensored int) error {
	query := "UPDATE TITLE " +
		"SET UNCENSORED = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, uncensored, titleID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to update UNCENSORED field for row with given ID from table TITLE: %w", err)
	}

	return nil
}

func (r *repositoryTitle) UpdateLangs(ctx context.Context, dbOps sqlite.DBOps, titleID string, langs string) error {
	query := "UPDATE TITLE " +
		"SET LANGS = ? " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, langs, titleID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to update LANGS field for row with given ID from table TITLE: %w", err)
	}

	return nil
}

func (r *repositoryTitle) DeleteAllByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	query := "DELETE FROM TITLE " +
		"WHERE LIBRARY_ID = ?"

	_, err := dbOps.ExecContext(ctx, query, libraryID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to delete rows with given LIBRARY_ID from table TITLE: %w", err)
	}

	return nil
}

func (r *repositoryTitle) DeleteByID(ctx context.Context, dbOps sqlite.DBOps, titleID string) error {
	query := "DELETE FROM TITLE " +
		"WHERE ID = ?"

	_, err := dbOps.ExecContext(ctx, query, titleID)
	if err != nil {
		return fmt.Errorf("rTitle - failed to delete row with given ID from table TITLE: %w", err)
	}

	return nil
}
