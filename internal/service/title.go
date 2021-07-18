package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/repository"
)

type ServiceTitle interface {
	CreateTitle(context.Context, sqlite.DBOps, *model.Title) error
	SearchTitles(context.Context, sqlite.DBOps, *model.TitleQuery) ([]*model.Title, error)
	GetTitleByID(context.Context, sqlite.DBOps, string) (*model.Title, error)
	GetTitlesByLibraryID(context.Context, sqlite.DBOps, string) ([]*model.Title, error)
	StreamTitleCoverByID(context.Context, sqlite.DBOps, io.Writer, string) error
	UpdateTitleModifiedTime(context.Context, sqlite.DBOps, string, string) error
	UpdateTitleCoverDimension(context.Context, sqlite.DBOps, string, int, int) error
	DeleteTitlesByLibraryID(context.Context, sqlite.DBOps, string) error
	DeleteTitleByID(context.Context, sqlite.DBOps, string) error
}

type serviceTitle struct {
	repositoryTitle repository.RepositoryTitle
	serviceBook     ServiceBook
}

func NewServiceTitle(rTitle repository.RepositoryTitle, sBook ServiceBook) ServiceTitle {
	return &serviceTitle{
		repositoryTitle: rTitle,
		serviceBook:     sBook,
	}
}

func (s *serviceTitle) CreateTitle(ctx context.Context, dbOps sqlite.DBOps, title *model.Title) error {
	err := s.repositoryTitle.Insert(ctx, dbOps, title)
	if err != nil {
		return fmt.Errorf("sTitle - failed to create title in DB: %w", err)
	}

	return nil
}

func (s *serviceTitle) SearchTitles(ctx context.Context, dbOps sqlite.DBOps, titleQuery *model.TitleQuery) ([]*model.Title, error) {
	titles, err := s.repositoryTitle.Find(ctx, dbOps, titleQuery)
	if err != nil {
		return nil, fmt.Errorf("sTitle - failed to search for titles with given queries in DB: %w", err)
	}

	return titles, nil
}

func (s *serviceTitle) GetTitleByID(ctx context.Context, dbOps sqlite.DBOps, titleID string) (*model.Title, error) {
	title, err := s.repositoryTitle.FindByID(ctx, dbOps, titleID)
	if err != nil {
		return nil, fmt.Errorf("sTitle - failed to find title with given ID in DB: %w", err)
	}

	return title, nil
}

func (s *serviceTitle) GetTitlesByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) ([]*model.Title, error) {
	titles, err := s.repositoryTitle.FindAllByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return nil, fmt.Errorf("sTitle - failed to find all titles with given library ID in DB: %w", err)
	}

	return titles, nil
}

func (s *serviceTitle) UpdateTitleModifiedTime(ctx context.Context, dbOps sqlite.DBOps, titleID string, modTime string) error {
	err := s.repositoryTitle.UpdateModifiedTime(ctx, dbOps, titleID, modTime)
	if err != nil {
		return fmt.Errorf("sTitle - failed to update title's modified time with given title ID in DB: %w", err)
	}

	return nil
}

func (s *serviceTitle) UpdateTitleCoverDimension(ctx context.Context, dbOps sqlite.DBOps, titleID string, coverWidth int, coverHeight int) error {
	err := s.repositoryTitle.UpdateCoverDimension(ctx, dbOps, titleID, coverWidth, coverHeight)
	if err != nil {
		return fmt.Errorf("sTitle - failed to update title's cover dimension with given title ID in DB: %w", err)
	}

	return nil
}

func (s *serviceTitle) DeleteTitlesByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	err := s.repositoryTitle.DeleteAllByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sTitle - failed to delete all titles with given library ID in DB: %w", err)
	}

	return nil
}

func (s *serviceTitle) DeleteTitleByID(ctx context.Context, dbOps sqlite.DBOps, titleID string) error {
	err := s.repositoryTitle.DeleteByID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("sTitle - failed to delete title by ID in DB: %w", err)
	}
	err = s.serviceBook.DeleteBooksByTitleID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("sTitle - failed to delete books of deleted title: %w", err)
	}

	return nil
}

func (s *serviceTitle) StreamTitleCoverByID(ctx context.Context, dbOps sqlite.DBOps, writer io.Writer, titleID string) error {
	title, err := s.repositoryTitle.FindByID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("sTitle - failed to find title with given ID in DB: %w", err)
	}
	titleCoverPath := filepath.Join(title.URL, "poster.jpg")
	titleCoverFile, err := os.Open(titleCoverPath)
	if err != nil {
		return fmt.Errorf("sTitle - failed to open cover file: %w", err)
	}

	io.Copy(writer, titleCoverFile)
	return nil
}
