package service

import (
	"context"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type libraryRepository interface {
	Insert(context.Context, sqlite.DBOps, *model.Library) error
	FindAll(context.Context, sqlite.DBOps) ([]*model.Library, error)
	Delete(context.Context, sqlite.DBOps, string) error
}

type LibraryService struct {
	libraryRepository libraryRepository
	db                sqlite.DB
}

func NewLibraryService(r libraryRepository, db sqlite.DB) *LibraryService {
	return &LibraryService{
		libraryRepository: r,
		db:                db,
	}
}

func (s *LibraryService) CreateLibrary(ctx context.Context, library *model.Library) error {
	return s.libraryRepository.Insert(ctx, s.db, library)
}

func (s *LibraryService) GetLibraries(ctx context.Context) ([]*model.Library, error) {
	return s.libraryRepository.FindAll(ctx, s.db)
}

func (s *LibraryService) DeleteLibrary(ctx context.Context, libraryID string) error {
	return s.libraryRepository.Delete(ctx, s.db, libraryID)
}
