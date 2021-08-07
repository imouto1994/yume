package service

import (
	"context"
	"fmt"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/repository"
	"go.uber.org/zap"
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
		return fmt.Errorf("sLibrary - failed to use service Title to delete titles of deleted library: %w", err)
	}
	err = s.serviceBook.DeleteBooksByLibraryID(ctx, dbOps, libraryID)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to use service Book to delete books of deleted library: %w", err)
	}

	return nil
}

func (s *serviceLibrary) ScanLibrary(ctx context.Context, dbOps sqlite.DBOps, library *model.Library) error {
	scanResult, err := s.serviceScanner.ScanLibraryRoot(library.Root)
	if err != nil {
		return fmt.Errorf("sLibrary - failed to use service Scanner to scan library: %w", err)
	}
	zap.L().Info("sLibrary - successfully scanned library files", zap.Int("numTitles", len(scanResult.TitleByTitleName)))

	// Get all titles in DB
	dbTitles, err := s.serviceTitle.GetTitlesByLibraryID(ctx, dbOps, fmt.Sprintf("%d", library.ID))
	if err != nil {
		return fmt.Errorf("sLibrary - failed to use service Title get all current titles in scanned library: %w", err)
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
				return fmt.Errorf("sLibrary - failed to use service Title to delete non-existing titles in scanned library: %w", err)
			}
			zap.L().Info("sLibrary - successfully removed non-existing title", zap.String("name", dbTitle.Name))
		}
	}

	numBooks := 0
	for _, books := range scanResult.BooksByTitleName {
		numBooks += len(books)
	}

	bookScanChannel := make(chan error, numBooks)
	for _, title := range scanResult.TitleByTitleName {
		if dbTitle, ok := dbTitleByTitleName[title.Name]; ok {
			if dbTitle.UpdatedAt != title.UpdatedAt {
				// Update title's modified time
				err := s.serviceTitle.UpdateTitleModifiedTime(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.UpdatedAt)
				if err != nil {
					return fmt.Errorf("sLibrary - failed to use service Title to update title's modified time from scanned library: %w", err)
				}

				// Update title's cover dimension if necessary
				if dbTitle.CoverHeight != title.CoverHeight || dbTitle.CoverWidth != title.CoverWidth {
					err = s.serviceTitle.UpdateTitleCoverDimension(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.CoverWidth, title.CoverHeight)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to use service Title to update title's cover dimension from scanned library: %w", err)
					}
				}

				// Update title's supported languages if necessary
				if dbTitle.Langs != title.Langs {
					err = s.serviceTitle.UpdateTitleLangs(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.Langs)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to use service Title to update title's supported langs from scanned library: %w", err)
					}
				}

				// Update title's book count if necessary
				if dbTitle.BookCount != title.BookCount {
					err = s.serviceTitle.UpdateTitleBookCount(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.BookCount)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to use service Title to update title's book count from scanned library: %w", err)
					}
				}

				// Update title's flags if necessary
				if dbTitle.Uncensored != title.Uncensored {
					err = s.serviceTitle.UpdateTitleUncensored(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.Uncensored)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to use service Title to update title's uncensored flag from scanned library: %w", err)
					}
				}

				if dbTitle.Waifu2x != title.Waifu2x {
					err = s.serviceTitle.UpdateTitleWaifu2x(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.Waifu2x)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to use service Title to update title's waifu2x flag from scanned library: %w", err)
					}
				}

				if dbTitle.Webp != title.Webp {
					err = s.serviceTitle.UpdateTitleWebp(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID), title.Webp)
					if err != nil {
						return fmt.Errorf("sLibrary - failed to use service Title to update title's webp flag from scanned library: %w", err)
					}
				}

				dbBooks, err := s.serviceBook.GetBooksByTitleID(ctx, dbOps, fmt.Sprintf("%d", dbTitle.ID))
				if err != nil {
					return fmt.Errorf("sLibrary - failed to use service Book to get all current stored books in updated title from scanned library: %w", err)
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
							return fmt.Errorf("sLibrary - failed to use service Book to delete non-existing books in updated title from scanned library: %w", err)
						}
					}
				}

				for _, book := range books {
					if dbBook, ok := dbBookByBookName[book.Name]; ok {
						if dbBook.UpdatedAt != book.UpdatedAt {
							// Update book's modified time
							err = s.serviceBook.UpdateBookModifiedTime(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID), book.UpdatedAt)
							if err != nil {
								return fmt.Errorf("sLibrary - failed to use service Book to update book's modified time in updated title from scanned library: %w", err)
							}

							// Update book's page count
							if dbBook.PageCount != book.PageCount {
								err = s.serviceBook.UpdateBookPageCount(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID), book.PageCount)
								if err != nil {
									return fmt.Errorf("sLibrary - failed to use service Book to update book's page count in updated title from scanned library: %w", err)
								}
							}

							// Update book's format
							if dbBook.Format != book.Format {
								err = s.serviceBook.UpdateBookFormat(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID), book.Format)
								if err != nil {
									return fmt.Errorf("sLibrary - failed to use service Book to update book's format in updated title from scanned library: %w", err)
								}
							}

							// Delete all pages from updated book in updated title
							err = s.serviceBook.DeleteBookPages(ctx, dbOps, fmt.Sprintf("%d", dbBook.ID))
							if err != nil {
								return fmt.Errorf("sLibrary - failed to use service Book to delete pages of updated book in updated title from scanned library: %w", err)
							}

							// Rescan all pages from updated book in updated title
							go func(b *model.Book) {
								err = s.serviceBook.ScanBook(ctx, dbOps, b)
								if err != nil {
									bookScanChannel <- fmt.Errorf("sLibrary - failed to use service Book to scan updated book for updated title from scanned library: %w", err)
								} else {
									bookScanChannel <- nil
								}
							}(dbBook)
						} else {
							bookScanChannel <- nil
						}
					} else {
						book.LibraryID = library.ID
						book.TitleID = dbTitle.ID

						// Create new book entry in updated title
						err = s.serviceBook.CreateBook(ctx, dbOps, book)
						if err != nil {
							return fmt.Errorf("sLibrary - failed to use service Book to create new book for updated title from scanned library: %w", err)
						}

						// Scan new book in updated title
						go func(b *model.Book) {
							err = s.serviceBook.ScanBook(ctx, dbOps, b)
							if err != nil {
								bookScanChannel <- fmt.Errorf("sLibrary - failed to use service Book to scan new book for updated title from scanned library: %w", err)
							} else {
								bookScanChannel <- nil
							}
						}(book)
					}
				}
				zap.L().Info("sLibrary - successfully updated modified title", zap.String("name", title.Name))
			} else {
				books := scanResult.BooksByTitleName[title.Name]
				for range books {
					bookScanChannel <- nil
				}
			}
		} else {
			title.LibraryID = library.ID

			// Create new title entry
			err = s.serviceTitle.CreateTitle(ctx, dbOps, title)
			if err != nil {
				return fmt.Errorf("sLibrary - failed to use service Title to create new title for scanned library: %w", err)
			}

			books := scanResult.BooksByTitleName[title.Name]
			for _, book := range books {
				book.LibraryID = library.ID
				book.TitleID = title.ID

				// Create new book entry for new title
				err = s.serviceBook.CreateBook(ctx, dbOps, book)
				if err != nil {
					return fmt.Errorf("sLibrary - failed to use service Book to create book for new title from scanned library: %w", err)
				}

				// Scan new book for new title
				go func(b *model.Book) {
					err = s.serviceBook.ScanBook(ctx, dbOps, b)
					if err != nil {
						bookScanChannel <- fmt.Errorf("sLibrary - failed to use service Book to scan book for new title from scanned library: %w", err)
					} else {
						bookScanChannel <- nil
					}
				}(book)
			}
			zap.L().Info("sLibrary - successfully added new title", zap.String("name", title.Name))
		}
	}

	for i := 0; i < numBooks; i++ {
		err := <-bookScanChannel
		if err != nil {
			return err
		}
	}

	return nil
}
