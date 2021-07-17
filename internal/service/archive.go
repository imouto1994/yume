package service

import (
	"archive/zip"
	"fmt"
	"io"
	"sort"

	"github.com/imouto1994/yume/internal/model"
)

type ServiceArchive interface {
	GetReader(string) (*zip.ReadCloser, error)
	GetFilesCount(string) (int, error)
	StreamFileByIndex(io.Writer, string, int) error
}

type serviceArchive struct {
}

func NewServiceArchive() ServiceArchive {
	return &serviceArchive{}
}

func (s *serviceArchive) GetReader(archivePath string) (*zip.ReadCloser, error) {
	return zip.OpenReader(archivePath)
}

func (s *serviceArchive) GetFilesCount(archivePath string) (int, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open archive: %w", err)
	}
	defer reader.Close()

	return len(reader.File), nil
}

func (s *serviceArchive) StreamFileByIndex(writer io.Writer, archivePath string, index int) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer reader.Close()

	if reader.File[index] == nil {
		return fmt.Errorf("%w: file at given index does not exist in the given archive", model.ErrNotFound)
	}
	sort.Slice(reader.File, func(i, j int) bool {
		return reader.File[i].Name < reader.File[j].Name
	})
	indexedFileReader, err := reader.File[index].Open()
	if err != nil {
		return fmt.Errorf("failed to open file at given index in archive: %w", err)
	}
	defer indexedFileReader.Close()

	io.Copy(writer, indexedFileReader)

	return nil
}
