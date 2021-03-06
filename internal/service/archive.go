package service

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"

	"github.com/imouto1994/yume/internal/model"
)

type ServiceArchive interface {
	GetFilesCount(string) (int, error)
	StreamFileByIndex(io.Writer, string, int) (string, error)
}

type serviceArchive struct {
}

func NewServiceArchive() ServiceArchive {
	return &serviceArchive{}
}

func (s *serviceArchive) GetFilesCount(archivePath string) (int, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return 0, fmt.Errorf("sArchive - failed to open archive: %w", err)
	}
	defer reader.Close()

	return len(reader.File), nil
}

func (s *serviceArchive) StreamFileByIndex(writer io.Writer, archivePath string, index int) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("sArchive - failed to open archive: %w", err)
	}
	defer reader.Close()

	if reader.File[index] == nil {
		return "", fmt.Errorf("sArchive - %w: file at given index does not exist in the given archive", model.ErrNotFound)
	}
	indexedFileReader, err := reader.File[index].Open()
	if err != nil {
		return "", fmt.Errorf("sArchive - failed to open file at given index in archive: %w", err)
	}
	defer indexedFileReader.Close()

	io.Copy(writer, indexedFileReader)

	return filepath.Ext(reader.File[index].Name), nil
}
