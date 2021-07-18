package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/repository"
	"go.uber.org/zap"
)

type ServiceBook interface {
	CreateBook(context.Context, sqlite.DBOps, *model.Book) error
	GetBooksByTitleID(context.Context, sqlite.DBOps, string) ([]*model.Book, error)
	GetBookByID(context.Context, sqlite.DBOps, string) (*model.Book, error)
	GetBookPages(context.Context, sqlite.DBOps, string) ([]*model.Page, error)
	StreamBookPageByID(context.Context, sqlite.DBOps, io.Writer, string, int) (string, error)
	ScanBook(context.Context, sqlite.DBOps, *model.Book)
}

type serviceBook struct {
	repositoryBook repository.RepositoryBook
	repositoryPage repository.RepositoryPage
	serviceArchive ServiceArchive
	serviceImage   ServiceImage
}

type indexedBookPage struct {
	Index int
	File  *zip.File
}

func NewServiceBook(rBook repository.RepositoryBook, rPage repository.RepositoryPage, sArchive ServiceArchive, sImage ServiceImage) ServiceBook {
	return &serviceBook{
		repositoryBook: rBook,
		repositoryPage: rPage,
		serviceArchive: sArchive,
		serviceImage:   sImage,
	}
}

func (s *serviceBook) CreateBook(ctx context.Context, dbOps sqlite.DBOps, book *model.Book) error {
	return s.repositoryBook.Insert(ctx, dbOps, book)
}

func (s *serviceBook) GetBooksByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) ([]*model.Book, error) {
	return s.repositoryBook.FindAllByTitleID(ctx, dbOps, titleID)
}

func (s *serviceBook) GetBookByID(ctx context.Context, dbOps sqlite.DBOps, bookID string) (*model.Book, error) {
	return s.repositoryBook.FindByID(ctx, dbOps, bookID)
}

func (s *serviceBook) GetBookPages(ctx context.Context, dbOps sqlite.DBOps, bookID string) ([]*model.Page, error) {
	return s.repositoryPage.FindAllByBookID(ctx, dbOps, bookID)
}

func (s *serviceBook) ScanBook(ctx context.Context, dbOps sqlite.DBOps, book *model.Book) {
	bookReader, err := s.serviceArchive.GetReader(book.URL)
	if err != nil {
		zap.L().Error("failed to open book archive", zap.Error(err))
	}
	defer bookReader.Close()

	indexedPageFiles := make([]*indexedBookPage, len(bookReader.File))
	for index, pageFile := range bookReader.File {
		indexedPageFiles[index] = &indexedBookPage{
			Index: index,
			File:  pageFile,
		}
	}
	sort.Slice(indexedPageFiles, func(i, j int) bool {
		return indexedPageFiles[i].File.Name < indexedPageFiles[j].File.Name
	})

	for number, indexedPageFile := range indexedPageFiles {
		fileReader, err := indexedPageFile.File.Open()
		if err != nil {
			zap.L().Error("failed to open page in book archive", zap.Error(err))
			continue
		}
		width, height, err := s.serviceImage.GetDimensions(fileReader)
		if err != nil {
			zap.L().Error("failed to get dimensions of book page", zap.Error(err))
			continue
		}
		page := &model.Page{
			Index:  indexedPageFile.Index,
			Number: number,
			BookID: book.ID,
			Width:  width,
			Height: height,
		}
		err = s.repositoryPage.Insert(ctx, dbOps, page)
		if err != nil {
			zap.L().Error("failed to create page", zap.Error(err))
		}
	}
}

func (s *serviceBook) StreamBookPageByID(ctx context.Context, dbOps sqlite.DBOps, writer io.Writer, bookID string, pageNumber int) (string, error) {
	book, err := s.repositoryBook.FindByID(ctx, dbOps, bookID)
	if err != nil {
		return "", fmt.Errorf("failed to find book with given ID: %w", err)
	}

	return s.serviceArchive.StreamFileByIndex(writer, book.URL, pageNumber)
}
