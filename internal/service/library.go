package service

import (
	"context"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/repository"
	"go.uber.org/zap"
)

type ServiceLibrary interface {
	CreateLibrary(context.Context, sqlite.DBOps, *model.Library) error
	GetLibraries(context.Context, sqlite.DBOps) ([]*model.Library, error)
	GetLibraryByID(context.Context, sqlite.DBOps, string) (*model.Library, error)
	DeleteLibrary(context.Context, sqlite.DBOps, string) error
	ScanLibrary(context.Context, sqlite.DBOps, *model.Library)
}

type serviceLibrary struct {
	repositoryLibrary repository.RepositoryLibrary
	serviceScanner    ServiceScanner
	serviceTitle      ServiceTitle
	serviceBook       ServiceBook
}

func NewServiceLibrary(rLibrary repository.RepositoryLibrary, sScanner ServiceScanner, sTitle ServiceTitle, sBook ServiceBook) ServiceLibrary {
	return &serviceLibrary{
		repositoryLibrary: rLibrary,
		serviceScanner:    sScanner,
		serviceTitle:      sTitle,
		serviceBook:       sBook,
	}
}

func (s *serviceLibrary) CreateLibrary(ctx context.Context, dbOps sqlite.DBOps, library *model.Library) error {
	return s.repositoryLibrary.Insert(ctx, dbOps, library)
}

func (s *serviceLibrary) GetLibraries(ctx context.Context, dbOps sqlite.DBOps) ([]*model.Library, error) {
	return s.repositoryLibrary.FindAll(ctx, dbOps)
}

func (s *serviceLibrary) GetLibraryByID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) (*model.Library, error) {
	return s.repositoryLibrary.FindByID(ctx, dbOps, libraryID)
}

func (s *serviceLibrary) DeleteLibrary(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	return s.repositoryLibrary.Delete(ctx, dbOps, libraryID)
}

func (s *serviceLibrary) ScanLibrary(ctx context.Context, dbOps sqlite.DBOps, library *model.Library) {
	scanResult, err := s.serviceScanner.ScanLibraryRoot(library.Root)
	if err != nil {
		zap.L().Error("failed to scan library", zap.Error(err))
	}

	// Create title entries
	for _, title := range scanResult.TitleByTitleName {
		title.LibraryID = library.ID
		err = s.serviceTitle.CreateTitle(ctx, dbOps, title)
		if err != nil {
			zap.L().Error("failed to create title", zap.Error(err))
		}
	}

	for titleName, books := range scanResult.BooksByTitleName {
		title := scanResult.TitleByTitleName[titleName]
		titleID := title.ID

		// Create book entries
		for _, book := range books {
			book.LibraryID = library.ID
			book.TitleID = titleID
			err = s.serviceBook.CreateBook(ctx, dbOps, book)
			if err != nil {
				zap.L().Error("failed to create book", zap.Error(err))
			} else {
				// Scan book archive and create page entries
				go s.serviceBook.ScanBook(ctx, dbOps, book)
			}
		}
	}
}
