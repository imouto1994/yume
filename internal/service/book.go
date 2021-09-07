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
)

type ServiceBook interface {
	CreateBook(context.Context, sqlite.DBOps, *model.Book) error
	GetBooksByTitleID(context.Context, sqlite.DBOps, string) ([]*model.Book, error)
	GetBookByID(context.Context, sqlite.DBOps, string) (*model.Book, error)
	GetBookPages(context.Context, sqlite.DBOps, string) ([]*model.Page, error)
	GetBookPreviews(context.Context, sqlite.DBOps, string) ([]*model.Preview, error)
	StreamBookPageByID(context.Context, sqlite.DBOps, io.Writer, string, int) (string, error)
	StreamBookPreviewByID(context.Context, sqlite.DBOps, io.Writer, string, int) (string, error)
	ScanBook(context.Context, sqlite.DBOps, *model.Book) error
	UpdateBookModifiedTime(context.Context, sqlite.DBOps, string, string) error
	UpdateBookPreviewInfo(context.Context, sqlite.DBOps, string, *string, *string) error
	UpdateBookPageCount(context.Context, sqlite.DBOps, string, int) error
	UpdateBookPageFavorite(context.Context, sqlite.DBOps, string, int, int) error
	DeleteBookByID(context.Context, sqlite.DBOps, string) error
	DeleteBooksByLibraryID(context.Context, sqlite.DBOps, string) error
	DeleteBooksByTitleID(context.Context, sqlite.DBOps, string) error
	DeleteBookPages(context.Context, sqlite.DBOps, string) error
	DeleteBookPreviews(context.Context, sqlite.DBOps, string) error
}

type serviceBook struct {
	repositoryBook    repository.RepositoryBook
	repositoryPage    repository.RepositoryPage
	repositoryPreview repository.RepositoryPreview
	serviceArchive    ServiceArchive
	serviceImage      ServiceImage
}

type indexedFile struct {
	Index int
	File  *zip.File
}

func NewServiceBook(rBook repository.RepositoryBook, rPage repository.RepositoryPage, rPreview repository.RepositoryPreview, sArchive ServiceArchive, sImage ServiceImage) ServiceBook {
	return &serviceBook{
		repositoryBook:    rBook,
		repositoryPage:    rPage,
		repositoryPreview: rPreview,
		serviceArchive:    sArchive,
		serviceImage:      sImage,
	}
}

func (s *serviceBook) CreateBook(ctx context.Context, dbOps sqlite.DBOps, book *model.Book) error {
	err := s.repositoryBook.Insert(ctx, dbOps, book)
	if err != nil {
		return fmt.Errorf("sBook - failed to create book in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) GetBooksByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) ([]*model.Book, error) {
	books, err := s.repositoryBook.FindAllByTitleID(ctx, dbOps, titleID)
	if err != nil {
		return nil, fmt.Errorf("sBook - failed to find all books by given title ID in DB: %w", err)
	}

	return books, nil
}

func (s *serviceBook) GetBookByID(ctx context.Context, dbOps sqlite.DBOps, bookID string) (*model.Book, error) {
	book, err := s.repositoryBook.FindByID(ctx, dbOps, bookID)
	if err != nil {
		return nil, fmt.Errorf("sBook - failed to find book by given ID in DB: %w", err)
	}

	return book, nil
}

func (s *serviceBook) GetBookPages(ctx context.Context, dbOps sqlite.DBOps, bookID string) ([]*model.Page, error) {
	pages, err := s.repositoryPage.FindAllByBookID(ctx, dbOps, bookID)
	if err != nil {
		return nil, fmt.Errorf("sBook - failed to find all pages of given book ID in DB: %w", err)
	}

	return pages, nil
}

func (s *serviceBook) GetBookPreviews(ctx context.Context, dbOps sqlite.DBOps, bookID string) ([]*model.Preview, error) {
	previews, err := s.repositoryPreview.FindAllByBookID(ctx, dbOps, bookID)
	if err != nil {
		return nil, fmt.Errorf("sBook - failed to find all previews of given book ID in DB: %w", err)
	}

	return previews, nil
}

func (s *serviceBook) ScanBook(ctx context.Context, dbOps sqlite.DBOps, book *model.Book) error {
	pagesReader, err := zip.OpenReader(book.URL)
	if err != nil {
		return fmt.Errorf("sBook - failed to open book archive: %w", err)
	}
	defer pagesReader.Close()

	indexedPageFiles := make([]*indexedFile, len(pagesReader.File))
	for index, pageFile := range pagesReader.File {
		indexedPageFiles[index] = &indexedFile{
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
			return fmt.Errorf("sBook - failed to open page in book archive: %w", err)
		}
		width, height, err := s.serviceImage.GetDimensions(fileReader)
		if err != nil {
			return fmt.Errorf("sBook - failed to use service Image to get dimensions of book page: %w", err)
		}
		page := &model.Page{
			Index:     indexedPageFile.Index,
			Number:    number,
			BookID:    book.ID,
			TitleID:   book.TitleID,
			LibraryID: book.LibraryID,
			Width:     width,
			Height:    height,
		}
		err = s.repositoryPage.Insert(ctx, dbOps, page)
		if err != nil {
			return fmt.Errorf("sBook - failed to create page in DB: %w", err)
		}
	}

	if book.PreviewURL != nil {
		previewsReader, err := zip.OpenReader(*book.PreviewURL)
		if err != nil {
			return fmt.Errorf("sBook - failed to open book previews archive file: %w", err)
		}
		defer previewsReader.Close()

		indexedPreviewFiles := make([]*indexedFile, len(previewsReader.File))
		for index, previewFile := range previewsReader.File {
			indexedPreviewFiles[index] = &indexedFile{
				Index: index,
				File:  previewFile,
			}
		}
		sort.Slice(indexedPreviewFiles, func(i, j int) bool {
			return indexedPreviewFiles[i].File.Name < indexedPreviewFiles[j].File.Name
		})

		for number, indexedPreviewFile := range indexedPreviewFiles {
			preview := &model.Preview{
				Index:     indexedPreviewFile.Index,
				Number:    number,
				BookID:    book.ID,
				TitleID:   book.TitleID,
				LibraryID: book.LibraryID,
			}
			err = s.repositoryPreview.Insert(ctx, dbOps, preview)
			if err != nil {
				return fmt.Errorf("sBook - failed to create preview in DB: %w", err)
			}
		}
	}

	return nil
}

func (s *serviceBook) StreamBookPageByID(ctx context.Context, dbOps sqlite.DBOps, writer io.Writer, bookID string, pageIndex int) (string, error) {
	book, err := s.repositoryBook.FindByID(ctx, dbOps, bookID)
	if err != nil {
		return "", fmt.Errorf("sBook - failed to find book with given ID in DB: %w", err)
	}

	extension, err := s.serviceArchive.StreamFileByIndex(writer, book.URL, pageIndex)
	if err != nil {
		return "", fmt.Errorf("sBook - failed to use service Archive to stream file by index: %w", err)
	}

	return extension, nil
}

func (s *serviceBook) StreamBookPreviewByID(ctx context.Context, dbOps sqlite.DBOps, writer io.Writer, bookID string, pageIndex int) (string, error) {
	book, err := s.repositoryBook.FindByID(ctx, dbOps, bookID)
	if err != nil {
		return "", fmt.Errorf("sBook - failed to find book with given ID in DB: %w", err)
	}

	if book.PreviewURL == nil {
		return "", fmt.Errorf("sBook- %w: book does not have previews", model.ErrNotFound)
	}

	extension, err := s.serviceArchive.StreamFileByIndex(writer, *book.PreviewURL, pageIndex)
	if err != nil {
		return "", fmt.Errorf("sBook - failed to use service Archive to stream file by index: %w", err)
	}

	return extension, nil
}

func (s *serviceBook) UpdateBookModifiedTime(ctx context.Context, dbOps sqlite.DBOps, bookID string, modTime string) error {
	err := s.repositoryBook.UpdateModifiedTime(ctx, dbOps, bookID, modTime)
	if err != nil {
		return fmt.Errorf("sBook - failed to update book's modified time with given book ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) UpdateBookPreviewInfo(ctx context.Context, dbOps sqlite.DBOps, bookID string, previewURL *string, previewUpdatedAt *string) error {
	err := s.repositoryBook.UpdatePreview(ctx, dbOps, bookID, previewURL, previewUpdatedAt)
	if err != nil {
		return fmt.Errorf("sBook - failed to update book's preview info with given book ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) UpdateBookPageCount(ctx context.Context, dbOps sqlite.DBOps, bookID string, pageCount int) error {
	err := s.repositoryBook.UpdatePageCount(ctx, dbOps, bookID, pageCount)
	if err != nil {
		return fmt.Errorf("sBook - failed to update book's page count with given book ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) UpdateBookPageFavorite(ctx context.Context, dbOps sqlite.DBOps, bookID string, pageNumber int, favorite int) error {
	err := s.repositoryPage.UpdateFavorite(ctx, dbOps, bookID, pageNumber, favorite)
	if err != nil {
		return fmt.Errorf("sBook - failed to update favoriate of book page with given book ID and page number in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) DeleteBookByID(ctx context.Context, dbOps sqlite.DBOps, bookID string) error {
	err := s.repositoryBook.DeleteByID(ctx, dbOps, bookID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book with given book ID in DB: %w", err)
	}
	err = s.repositoryPage.DeleteAllByBookID(ctx, dbOps, bookID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book pages with given book ID in DB: %w", err)
	}
	err = s.repositoryPreview.DeleteAllByBookID(ctx, dbOps, bookID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book previews with given book ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) DeleteBooksByTitleID(ctx context.Context, dbOps sqlite.DBOps, titleID string) error {
	err := s.repositoryBook.DeleteAllByTitleID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete books with given title ID in DB: %w", err)
	}
	err = s.repositoryPage.DeleteAllByTitleID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book pages with given title ID in DB: %w", err)
	}
	err = s.repositoryPreview.DeleteAllByTitleID(ctx, dbOps, titleID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book previews with given title ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) DeleteBooksByLibraryID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	err := s.repositoryBook.DeleteAllByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete books with given library ID in DB: %w", err)
	}
	err = s.repositoryPage.DeleteAllByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book pages with given library ID in DB: %w", err)
	}
	err = s.repositoryPreview.DeleteAllByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book previews with given library ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) DeleteBookPages(ctx context.Context, dbOps sqlite.DBOps, bookID string) error {
	err := s.repositoryPage.DeleteAllByBookID(ctx, dbOps, bookID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book pages with given book ID in DB: %w", err)
	}

	return nil
}

func (s *serviceBook) DeleteBookPreviews(ctx context.Context, dbOps sqlite.DBOps, bookID string) error {
	err := s.repositoryPreview.DeleteAllByBookID(ctx, dbOps, bookID)
	if err != nil {
		return fmt.Errorf("sBook - failed to delete book previews with given book ID in DB: %w", err)
	}

	return nil
}
