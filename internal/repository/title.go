package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type RepositoryTitle interface {
	Insert(context.Context, sqlite.DBOps, *model.Title) error
	Find(context.Context, sqlite.DBOps, *model.TitleQuery) ([]*model.Title, error)
	FindAllByLibraryID(context.Context, sqlite.DBOps, string) ([]*model.Title, error)
	FindByID(context.Context, sqlite.DBOps, string) (*model.Title, error)
}

type repositoryTitle struct {
}

func NewRepositoryTitle() RepositoryTitle {
	return &repositoryTitle{}
}

func (r *repositoryTitle) Insert(ctx context.Context, db sqlite.DBOps, title *model.Title) error {
	query := "INSERT INTO TITLE (NAME, URL, CREATED_AT, UPDATED_AT, COVER_WIDTH, COVER_HEIGHT, LIBRARY_ID) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?)"

	result, err := db.ExecContext(ctx, query, title.Name, title.URL, title.CreatedAt, title.UpdatedAt, title.CoverWidth, title.CoverHeight, title.LibraryID)
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

func (r *repositoryTitle) Find(ctx context.Context, dbOps sqlite.DBOps, titleQuery *model.TitleQuery) ([]*model.Title, error) {
	var whereClauseBuilder strings.Builder
	if len(titleQuery.LibraryIDs) > 0 {
		if whereClauseBuilder.Len() > 0 {
			whereClauseBuilder.WriteString("AND ")
		}
		whereClauseBuilder.WriteString(fmt.Sprintf("LIBRARY_ID IN (%s) ", strings.Join(titleQuery.LibraryIDs, ",")))
	}
	if titleQuery.Search != "" {
		if whereClauseBuilder.Len() > 0 {
			whereClauseBuilder.WriteString("AND ")
		}
		whereClauseBuilder.WriteString(fmt.Sprintf("NAME LIKE '%%%s%%' ", titleQuery.Search))
	}

	orderByClause := "NAME ASC "
	if titleQuery.Sort == "created_at" {
		orderByClause = "CREATED_AT DESC "
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT * FROM TITLE ")
	if whereClauseBuilder.Len() > 0 {
		queryBuilder.WriteString(fmt.Sprintf("WHERE %s ", whereClauseBuilder.String()))
	}
	queryBuilder.WriteString(fmt.Sprintf("ORDER BY %s ", orderByClause))
	queryBuilder.WriteString(fmt.Sprintf("LIMIT %d ", titleQuery.Size))
	queryBuilder.WriteString(fmt.Sprintf("OFFSET %d ", titleQuery.Page*titleQuery.Size))

	titles := []*model.Title{}

	err := dbOps.SelectContext(ctx, &titles, queryBuilder.String())
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: no matched rows with specific LIBRARY_ID from table TITLE", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find rows with specific LIBRARY_ID from table TITLE: %w", err)
	}

	return titles, nil
}

func (r *repositoryTitle) FindAllByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) ([]*model.Title, error) {
	query := "SELECT * FROM TITLE " +
		"WHERE LIBRARY_ID = ?"

	titles := []*model.Title{}

	err := dbOps.SelectContext(ctx, &titles, query, libraryID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: no matched rows with specific LIBRARY_ID from table TITLE", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find rows with specific LIBRARY_ID from table TITLE: %w", err)
	}

	return titles, nil
}

func (r *repositoryTitle) FindByID(ctx context.Context, dbOps sqlite.DBOps, titleID string) (*model.Title, error) {
	query := "SELECT * FROM TITLE " +
		"WHERE ID = ?"

	title := model.Title{}

	err := dbOps.GetContext(ctx, &title, query, titleID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: no matched rows with specific ID from table TITLE", model.ErrNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("failed to find row with specific ID from table TITLE: %w", err)
	}

	return &title, nil
}
