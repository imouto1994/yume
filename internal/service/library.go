package service

import (
	"github.com/imouto1994/yume/internal/model"
)

type libraryRepository interface {
	Add(library *model.Library) error
}

type LibraryService struct {
	libraryRepository libraryRepository
}

func NewLibraryService(r libraryRepository) *LibraryService {
	return &LibraryService{
		libraryRepository: r,
	}
}

func (s *LibraryService) CreateLibrary(library *model.Library) error {
	return s.libraryRepository.Add(library)
}

func (s *LibraryService) GetLibraries() ([]*model.Library, error) {
	return nil, nil
}
