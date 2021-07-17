package service

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

func NewServiceScanner(sImage ServiceImage, sArchive ServiceArchive) ServiceScanner {
	return &serviceScanner{
		serviceArchive: sArchive,
		serviceImage:   sImage,
	}
}

func (s *serviceScanner) ScanLibraryRoot(libraryPath string) (*model.ScanResult, error) {
	files, err := os.ReadDir(libraryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read library root folder: %w", err)
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
		titleLastModifiedTime := folderInfo.ModTime().UTC().Format(time.RFC3339)

		title := &model.Title{
			Name:      titleName,
			URL:       filepath.Join(libraryPath, titleName),
			CreatedAt: titleLastModifiedTime,
			UpdatedAt: titleLastModifiedTime,
		}
		titleByTitleName[titleName] = title
	}

	// Scan title covers
	for _, titleFolder := range titleFolders {
		titleName := titleFolder.Name()
		width, height, err := s.scanTitleCover(filepath.Join(libraryPath, titleName))
		if err != nil {
			zap.L().Error("failed to scan title cover", zap.Error(err))
		}
		title := titleByTitleName[titleFolder.Name()]
		title.CoverWidth = width
		title.CoverHeight = height
	}

	// Scan books in each title
	booksByTitleName := make(map[string][]*model.Book)
	for _, titleFolder := range titleFolders {
		books := s.scanTitleFolder(filepath.Join(libraryPath, titleFolder.Name()))
		if books != nil {
			booksByTitleName[titleFolder.Name()] = books
		}
	}

	return &model.ScanResult{
		TitleByTitleName: titleByTitleName,
		BooksByTitleName: booksByTitleName,
	}, nil
}

func (s *serviceScanner) scanTitleCover(titleFolderPath string) (int, int, error) {
	titleCoverPath := filepath.Join(titleFolderPath, "cover.png")
	titleCoverFile, err := os.Open(titleCoverPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open cover file: %w", err)
	}

	return s.serviceImage.GetDimensions(titleCoverFile)
}

func (s *serviceScanner) scanTitleFolder(titleFolderPath string) []*model.Book {
	files, err := os.ReadDir(titleFolderPath)
	if err != nil {
		zap.L().Error("failed to read title folder", zap.Error(err))
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
