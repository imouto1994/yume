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
	GetTitlesByLibrary(context.Context, sqlite.DBOps, string) ([]*model.Title, error)
	GetTitleByID(context.Context, sqlite.DBOps, string) (*model.Title, error)
	StreamTitleCoverByID(context.Context, sqlite.DBOps, io.Writer, string) error
}

type serviceTitle struct {
	repositoryTitle repository.RepositoryTitle
}

func NewServiceTitle(rTitle repository.RepositoryTitle) ServiceTitle {
	return &serviceTitle{
		repositoryTitle: rTitle,
	}
}

func (s *serviceTitle) CreateTitle(ctx context.Context, dbOps sqlite.DBOps, title *model.Title) error {
	return s.repositoryTitle.Insert(ctx, dbOps, title)
}

func (s *serviceTitle) GetTitlesByLibrary(ctx context.Context, dbOps sqlite.DBOps, libraryID string) ([]*model.Title, error) {
	return nil, nil
}

func (s *serviceTitle) GetTitleByID(ctx context.Context, dbOps sqlite.DBOps, titleID string) (*model.Title, error) {
	return nil, nil
}

func (s *serviceTitle) StreamTitleCoverByID(ctx context.Context, dbOps sqlite.DBOps, writer io.Writer, titleID string) error {
	title, err := s.repositoryTitle.FindByID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("failed to find title with given ID: %w", err)
	}
	titleCoverPath := filepath.Join(title.URL, "cover.png")
	titleCoverFile, err := os.Open(titleCoverPath)
	if err != nil {
		return fmt.Errorf("failed to open cover file: %w", err)
	}

	io.Copy(writer, titleCoverFile)
	return nil
}
