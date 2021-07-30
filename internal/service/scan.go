package service

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/imouto1994/yume/internal/model"
	"go.uber.org/zap"
)

type ServiceScanner interface {
	ScanLibraryRoot(string) (*model.ScanResult, error)
}

type serviceScanner struct {
	serviceArchive ServiceArchive
	serviceImage   ServiceImage
}

type titleCoverScanResult struct {
	Name   string
	Width  int
	Height int
}

type titleBooksScanResult struct {
	Name  string
	Books []*model.Book
}

func NewServiceScanner(sImage ServiceImage, sArchive ServiceArchive) ServiceScanner {
	return &serviceScanner{
		serviceArchive: sArchive,
		serviceImage:   sImage,
	}
}

func (s *serviceScanner) ScanLibraryRoot(libraryPath string) (*model.ScanResult, error) {
	files, err := os.ReadDir(libraryPath)
	if err != nil {
		return nil, fmt.Errorf("sScan - failed to read library root folder: %w", err)
	}

	// Filter for titles
	titleFolders := []fs.DirEntry{}
	for _, file := range files {
		fileInfo, _ := file.Info()
		if fileInfo.IsDir() {
			titleFolders = append(titleFolders, file)
		}
	}

	// Create titles
	titleByTitleName := make(map[string]*model.Title)
	for _, titleFolder := range titleFolders {
		titleName := titleFolder.Name()
		folderInfo, _ := titleFolder.Info()
		folderLastModifiedTime := folderInfo.ModTime().UTC().Format(time.RFC3339)
		titleCreatedTime := folderLastModifiedTime

		// Check for specified created time
		openedCurlyLastIndex := strings.LastIndex(titleName, "{")
		closedCurlyLastIndex := strings.LastIndex(titleName, "}")
		if openedCurlyLastIndex != -1 && closedCurlyLastIndex != -1 && closedCurlyLastIndex > openedCurlyLastIndex {
			specifiedTimeSubstring := titleName[(openedCurlyLastIndex + 1):closedCurlyLastIndex]
			specifiedTime, err := time.Parse("2006-01-02", specifiedTimeSubstring)
			if err == nil {
				titleCreatedTime = specifiedTime.UTC().Format(time.RFC3339)
			}
		}

		title := &model.Title{
			Name:      titleName,
			URL:       filepath.Join(libraryPath, titleName),
			CreatedAt: titleCreatedTime,
			UpdatedAt: folderLastModifiedTime,
		}
		titleByTitleName[titleName] = title
	}

	// Scan title covers
	coverScanChannel := make(chan *titleCoverScanResult, len(titleFolders))

	for _, titleFolder := range titleFolders {
		titleName := titleFolder.Name()
		go func(name string) {
			width, height, err := s.scanTitleCover(filepath.Join(libraryPath, name))
			if err != nil {
				zap.L().Error("sScan - failed to scan title cover", zap.Error(err))
				coverScanChannel <- nil
			}
			coverScanChannel <- &titleCoverScanResult{
				Name:   name,
				Width:  width,
				Height: height,
			}
		}(titleName)
	}

	for range titleFolders {
		coverScanResult := <-coverScanChannel
		if coverScanResult != nil {
			title := titleByTitleName[coverScanResult.Name]
			title.CoverWidth = coverScanResult.Width
			title.CoverHeight = coverScanResult.Height
		}
	}

	// Scan books in each title
	booksScanChannel := make(chan *titleBooksScanResult, len(titleFolders))
	booksByTitleName := make(map[string][]*model.Book)

	for _, titleFolder := range titleFolders {
		titleName := titleFolder.Name()
		go func(name string) {
			books := s.scanTitleFolder(filepath.Join(libraryPath, name))
			booksScanChannel <- &titleBooksScanResult{
				Name:  name,
				Books: books,
			}
		}(titleName)
	}

	for range titleFolders {
		booksScanResult := <-booksScanChannel
		titleBooks := booksScanResult.Books
		title := titleByTitleName[booksScanResult.Name]

		// Set number of books in title
		title.BookCount = len(titleBooks)

		// Set flags for title
		for _, book := range titleBooks {
			bookName := book.Name
			if strings.HasSuffix(bookName, "]") {
				openedBracketIndex := strings.LastIndex(bookName, "[")
				if openedBracketIndex > 0 {
					flags := strings.Split(bookName[(openedBracketIndex+1):(len(bookName)-1)], ",")
					for _, flag := range flags {
						flag = strings.TrimSpace(flag)
						if flag == "Uncensored" || flag == "Decensored" {
							title.Uncensored = 1
						} else if flag == "Waifu2x" {
							title.Waifu2x = 1
						}
					}

				}
			}
		}

		// Scan title's supported languages
		langsSet := make(map[string]bool)
		for _, book := range titleBooks {
			bookName := book.Name
			openedBracketIndex := strings.Index(bookName, "[")
			closedBracketIndex := strings.Index(bookName, "]")
			if openedBracketIndex == 0 && closedBracketIndex != -1 {
				lang := bookName[(openedBracketIndex + 1):closedBracketIndex]
				langsSet[lang] = true
			} else {
				langsSet["jp"] = true
			}
		}
		langs := []string{}
		for lang := range langsSet {
			langs = append(langs, lang)
		}
		sort.Strings(langs)
		title.Langs = strings.Join(langs, ",")

		booksByTitleName[booksScanResult.Name] = booksScanResult.Books
	}

	return &model.ScanResult{
		TitleByTitleName: titleByTitleName,
		BooksByTitleName: booksByTitleName,
	}, nil
}

func (s *serviceScanner) scanTitleCover(titleFolderPath string) (int, int, error) {
	titleCoverPath := filepath.Join(titleFolderPath, "poster.jpg")
	titleCoverFile, err := os.Open(titleCoverPath)
	if err != nil {
		return 0, 0, fmt.Errorf("sScan - failed to open cover file: %w", err)
	}

	return s.serviceImage.GetDimensions(titleCoverFile)
}

func (s *serviceScanner) scanTitleFolder(titleFolderPath string) []*model.Book {
	files, err := os.ReadDir(titleFolderPath)
	if err != nil {
		zap.L().Error("sScan - failed to read title folder", zap.Error(err))
		return nil
	}

	books := []*model.Book{}
	for _, file := range files {
		fileInfo, _ := file.Info()
		if fileInfo.IsDir() {
			continue
		}

		fileName := file.Name()
		fileExtension := filepath.Ext(fileName)
		if fileExtension == ".cbz" {
			bookFilePath := filepath.Join(titleFolderPath, fileName)
			bookLastModifiedTime := fileInfo.ModTime().UTC().Format(time.RFC3339)
			pageCount, _ := s.serviceArchive.GetFilesCount(bookFilePath)

			book := &model.Book{
				Name:      strings.TrimSuffix(fileName, fileExtension),
				URL:       bookFilePath,
				CreatedAt: bookLastModifiedTime,
				UpdatedAt: bookLastModifiedTime,
				PageCount: pageCount,
			}
			books = append(books, book)
		}
	}

	return books
}
