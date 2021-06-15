package service

import (
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/repository"
	"go.uber.org/zap"
)

type LibraryService struct {
	libraryRepository *repository.LibraryRepository
	logger            *zap.Logger
}

func NewLibraryService(r *repository.LibraryRepository, l *zap.Logger) *LibraryService {
	return &LibraryService{
		libraryRepository: r,
		logger:            l,
	}
}

func (s *LibraryService) CreateLibrary() error {
	return nil
}

func (s *LibraryService) GetLibraries() ([]*model.Library, error) {
	return nil, nil
}
