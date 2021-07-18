package service

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/repository"
)

type ServiceLibrary interface {
	CreateLibrary(context.Context, sqlite.DBOps, *model.Library) error
	GetLibraries(context.Context, sqlite.DBOps) ([]*model.Library, error)
	GetLibraryByID(context.Context, sqlite.DBOps, string) (*model.Library, error)
	DeleteLibraryByID(context.Context, sqlite.DBOps, string) error
	ScanLibrary(context.Context, sqlite.DBOps, *model.Library) error
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
	err := s.repositoryLibrary.Insert(ctx, dbOps, library)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to create library in DB: %w", err)
	}

	return nil
}

func (s *serviceLibrary) GetLibraries(ctx context.Context, dbOps sqlite.DBOps) ([]*model.Library, error) {
	libraries, err := s.repositoryLibrary.FindAll(ctx, dbOps)
	if err != nil {
		return nil, fmt.Errorf("sLibrary - failed to get all libraries in DB: %w", err)
	}

	return libraries, nil
}

func (s *serviceLibrary) GetLibraryByID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) (*model.Library, error) {
	library, err := s.repositoryLibrary.FindByID(ctx, dbOps, libraryID)
	if err != nil {
		return nil, fmt.Errorf("sLibrary - failed to library by ID in DB: %w", err)
	}

	return library, nil
}

func (s *serviceLibrary) DeleteLibraryByID(ctx context.Context, dbOps sqlite.DBOps, libraryID string) error {
	err := s.repositoryLibrary.DeleteByID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to delete library by ID in DB: %w", err)
	}
	err = s.serviceTitle.DeleteTitlesByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to delete titles of deleted library: %w", err)
	}
	err = s.serviceBook.DeleteBooksByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to delete books of deleted library: %w", err)
	}

	return nil
}

func (s *serviceLibrary) ScanLibrary(ctx context.Context, dbOps sqlite.DBOps, library *model.Library) error {
	scanResult, err := s.serviceScanner.ScanLibraryRoot(library.Root)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to scan library: %w", err)
	}

	// Get all titles in DB
	dbTitles, err := s.serviceTitle.GetTitlesByLibraryID(ctx, dbOps, fmt.Sprintf("%d", library.ID))
	if err != nil {
		return fmt.Errorf("sLibrary - failed to get all current stored titles in scanned library: %w", err)
	}
	dbTitleByTitleName := make(map[string]*model.Title)
	for _, dbTitle := range dbTitles {
		dbTitleByTitleName[dbTitle.Name] = dbTitle
	}

	// Remove titles not existing anymore
	for _, dbTitle := range dbTitles {
		if _, ok := scanResult.TitleByTitleName[dbTitle.Name]; !ok {
			err = s.serviceTitle.DeleteTitleByID(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID))
			if err != nil {
				return fmt.Errorf("sLibrary - failed to delete non-existing titles in scanned library: %w", err)
			}
		}
	}

	for _, title := range scanResult.TitleByTitleName {
		if dbTitle, ok := dbTitleByTitleName[title.Name]; ok {
			if dbTitle.UpdatedAt != title.UpdatedAt {
				// Update title's modified time
				err := s.serviceTitle.UpdateTitleModifiedTime(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.UpdatedAt)
				if err != nil {
					return fmt.Errorf("sLibrary - failed to update title's modified time from scanned library: %w", err)
				}

				// Update title's cover dimension
				if dbTitle.CoverHeight != title.CoverHeight || dbTitle.CoverWidth != title.CoverWidth {
					err = s.serviceTitle.UpdateTitleCoverDimension(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.CoverWidth, title.CoverHeight)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to update title's cover dimension from scanned library: %w", err)
					}
				}

				dbBooks, err := s.serviceBook.GetBooksByTitleID(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID))
				if err != nil {
					return fmt.Errorf("sLibrary - failed to get all current stored books in updated title from scanned library: %w", err)
				}
				books := scanResult.BooksByTitleName[title.Name]
				dbBookByBookName := make(map[string]*model.Book)
				bookByBookName := make(map[string]*model.Book)
				for _, dbBook := range dbBooks {
					dbBookByBookName[dbBook.Name] = dbBook
				}
				for _, book := range books {
					bookByBookName[book.Name] = book
				}

				for _, dbBook := range dbBooks {
					if _, ok := bookByBookName[dbBook.Name]; !ok {
						// Remove book not existing anymore
						err = s.serviceBook.DeleteBookByID(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID))
						if err != nil {
							return fmt.Errorf("sLibrary - failed to delete non-existing books in updated title from scanned library: %w", err)
						}
					}
				}

				for _, book := range books {
					if dbBook, ok := dbBookByBookName[book.Name]; ok {
						if dbBook.UpdatedAt != book.UpdatedAt {
							// Update book's modified time
							err = s.serviceBook.UpdateBookModifiedTime(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID), book.UpdatedAt)
							if err != nil {
								return fmt.Errorf("sLibrary - failed to update book's modified time in updated title from scanned library: %w", err)
							}

							// Update book's page count
							if dbBook.PageCount != book.PageCount {
								err = s.serviceBook.UpdateBookPageCount(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID), book.PageCount)
								if err != nil {
									return fmt.Errorf("sLibrary - failed to update book's page count in updated title from scanned library: %w", err)
								}
							}

							// Delete all pages from updated book in updated title
							err = s.serviceBook.DeleteBookPages(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID))
							if err != nil {
								return fmt.Errorf("sLibrary - failed to delete pages of updated book in updated title from scanned library: %w", err)
							}

							// Rescan all pages from updated book in updated title
							err = s.serviceBook.ScanBook(ctx, dbOps, dbBook)
							if err != nil {
								return fmt.Errorf("sLibrary - failed to scan updated book for updated title from scanned library: %w", err)
							}
						} else {
							book.LibraryID = library.ID
							book.TitleID = dbTitle.ID

							// Create new book entry in updated title
							err = s.serviceBook.CreateBook(ctx, dbOps, book)
							if err != nil {
								return fmt.Errorf("sLibrary - failed to create new book for updated title from scanned library: %w", err)
							}

							// Scan new book in updated title
							err = s.serviceBook.ScanBook(ctx, dbOps, book)
							if err != nil {
								return fmt.Errorf("sLibrary - failed to scan new book for updated title from scanned library: %w", err)
							}
						}
					}
				}
			}
		} else {
			title.LibraryID = library.ID

			// Create new title entry
			err = s.serviceTitle.CreateTitle(ctx, dbOps, title)
			if err != nil {
				return fmt.Errorf("sLibrary - failed to create new title for scanned library: %w", err)
			}

			books := scanResult.BooksByTitleName[title.Name]
			for _, book := range books {
				book.LibraryID = library.ID
				book.TitleID = title.ID

				// Create new book entry for new title
				err = s.serviceBook.CreateBook(ctx, dbOps, book)
				if err != nil {
					return fmt.Errorf("sLibrary - failed to create book for new title from scanned library: %w", err)
				}

				// Scan new book for new title
				err = s.serviceBook.ScanBook(ctx, dbOps, book)
				if err != nil {
					return fmt.Errorf("sLibrary - failed to scan book for new title from scanned library: %w", err)
				}
			}
		}
	}

	return nil
}
