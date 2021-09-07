package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
)

type ServiceSubtitle interface {
	CreateSubtitle(context.Context, sqlite.DBOps, *model.Subtitle) error
}

type serviceSubtitle struct {
	serviceLibrary ServiceLibrary
	serviceBook    ServiceBook
	serviceTitle   ServiceTitle
}

func NewServiceSubtitle(sLibrary ServiceLibrary, sBook ServiceBook, sTitle ServiceTitle) ServiceSubtitle {
	return &serviceSubtitle{
		serviceLibrary: sLibrary,
		serviceBook:    sBook,
		serviceTitle:   sTitle,
	}
}

func (s *serviceSubtitle) CreateSubtitle(ctx context.Context, dbOps sqlite.DBOps, subtitle *model.Subtitle) error {
	library, err := s.serviceLibrary.GetLibraryByID(ctx, dbOps, subtitle.LibraryID)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to use service Library to get library of subtitle: %w", err)
	}
	book, err := s.serviceBook.GetBookByID(ctx, dbOps, subtitle.BookID)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to use service Book to get book of subtitle: %w", err)
	}
	title, err := s.serviceTitle.GetTitleByID(ctx, dbOps, fmt.Sprintf("%d", book.TitleID))
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to use service Title to get title of subtitle: %w", err)
	}

	libraryFolder := library.Root

	// Determine subtitle name
	titleName := title.Name
	titleMain := strings.TrimSpace(titleName[(strings.Index(titleName, "]") + 1):strings.LastIndex(titleName, "{")])
	titleDate := titleName[(strings.LastIndex(titleName, "{") + 1):strings.LastIndex(titleName, "}")]
	subtitleName := fmt.Sprintf("[%s (%s)] %s {%s}", subtitle.Author, titleMain, subtitle.Name, titleDate)

	// Create subtitle folder
	subtitleFolderPath := filepath.Join(libraryFolder, subtitleName)
	os.MkdirAll(subtitleFolderPath, os.ModePerm)

	// Determine subtitle book lang
	bookLang := "en"
	if strings.HasPrefix(book.Name, "[JP]") {
		bookLang = "jp"
	}

	// Determine subtitle book tags
	tagsString := ""
	if strings.HasSuffix(book.Name, "]") {
		tagsString = book.Name[(strings.LastIndex(book.Name, "[") + 1):strings.LastIndex(book.Name, "]")]
	}

	// Determine subtitle book name
	subtitleBookName := fmt.Sprintf("[%s] %s", strings.ToUpper(bookLang), subtitle.Name)
	if tagsString != "" {
		subtitleBookName = fmt.Sprintf("%s [%s]", subtitleBookName, tagsString)
	}
	var coverFilePath string

	// Create subtitle book
	subtitleBookPath := filepath.Join(subtitleFolderPath, fmt.Sprintf("%s.cbz", subtitleBookName))
	subtitleBookZipFile, err := os.Create(subtitleBookPath)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to initialize zip file of subtitle book: %w", err)
	}
	subtitleBookZipWriter := zip.NewWriter(subtitleBookZipFile)

	pagesReader, err := zip.OpenReader(book.URL)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to open book archive: %w", err)
	}
	defer pagesReader.Close()
	pageFiles := append([]*zip.File(nil), pagesReader.File...)
	sort.Slice(pageFiles, func(i, j int) bool {
		return pageFiles[i].Name < pageFiles[j].Name
	})
	for i := subtitle.PageStartNumber; i <= subtitle.PageEndNumber; i++ {
		pageFile := pageFiles[i]
		err := addFileToZip(subtitleBookZipWriter, pageFile)
		if err != nil {
			return fmt.Errorf("sSubtitle - failed to add page file to subtitle book archive: %w", err)
		}
	}
	subtitleBookZipWriter.Close()
	subtitleBookZipFile.Close()

	// Create subtitle book backup
	backupPath := filepath.Join(filepath.Dir(book.URL), fmt.Sprintf("%s - Backup.zip", book.Name))
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("sSubtitle - backup file is not available for book: %w", err)
	}

	subtitleBookBackupPath := filepath.Join(subtitleFolderPath, fmt.Sprintf("%s - Backup.zip", subtitleBookName))
	subtitleBookBackupZipFile, err := os.Create(subtitleBookBackupPath)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to initialize zip file of subtitle book backup: %w", err)
	}
	subtitleBookBackupZipWriter := zip.NewWriter(subtitleBookBackupZipFile)

	backupPagesReader, err := zip.OpenReader(backupPath)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to open book backup archive: %w", err)
	}
	defer backupPagesReader.Close()
	backupPageFiles := append([]*zip.File(nil), backupPagesReader.File...)
	sort.Slice(backupPageFiles, func(i, j int) bool {
		return backupPageFiles[i].Name < backupPageFiles[j].Name
	})
	for i := subtitle.PageStartNumber; i <= subtitle.PageEndNumber; i++ {
		backupPageFile := backupPageFiles[i]

		if i == subtitle.PageStartNumber {
			// Extract first image in backup file to create cover thumbnail for subtitle book
			coverFilePath = filepath.Join(subtitleFolderPath, backupPageFile.Name)
			coverFileWriter, err := os.OpenFile(coverFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, backupPageFile.Mode())
			if err != nil {
				return fmt.Errorf("sSubtitle - failed to extract cover file from backup archive: %w", err)
			}
			backupPageFileRead, err := backupPageFile.Open()
			if err != nil {
				return fmt.Errorf("sSubtitle - failed to open backup page file used for cover: %w", err)
			}
			_, err = io.Copy(coverFileWriter, backupPageFileRead)
			coverFileWriter.Close()
			backupPageFileRead.Close()
			if err != nil {
				return fmt.Errorf("sSubtitle - failed to extract cover file from backup archive: %w", err)
			}
		}

		err := addFileToZip(subtitleBookBackupZipWriter, backupPageFile)
		if err != nil {
			return fmt.Errorf("sSubtitle - failed to add page file to subtitle book backup archive: %w", err)
		}
	}
	subtitleBookBackupZipWriter.Close()
	subtitleBookBackupZipFile.Close()

	// Create subtitle book preview
	previewPath := filepath.Join(filepath.Dir(book.URL), fmt.Sprintf("%s - Preview.zip", book.Name))
	if _, err := os.Stat(previewPath); err != nil {
		return fmt.Errorf("sSubtitle - preview file is not available for book: %w", err)
	}

	subtitleBookPreviewPath := filepath.Join(subtitleFolderPath, fmt.Sprintf("%s - Preview.zip", subtitleBookName))
	subtitleBookPreviewZipFile, err := os.Create(subtitleBookPreviewPath)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to initialize zip file of subtitle book preview: %w", err)
	}
	subtitleBookPreviewZipWriter := zip.NewWriter(subtitleBookPreviewZipFile)

	previewPagesReader, err := zip.OpenReader(previewPath)
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to open book preview archive: %w", err)
	}
	defer previewPagesReader.Close()
	previewPageFiles := append([]*zip.File(nil), previewPagesReader.File...)
	sort.Slice(previewPageFiles, func(i, j int) bool {
		return previewPageFiles[i].Name < previewPageFiles[j].Name
	})
	for i := subtitle.PageStartNumber; i <= subtitle.PageEndNumber; i++ {
		previewPageFile := previewPageFiles[i]
		err := addFileToZip(subtitleBookPreviewZipWriter, previewPageFile)
		if err != nil {
			return fmt.Errorf("sSubtitle - failed to add page file to subtitle book preview archive: %w", err)
		}
	}
	subtitleBookPreviewZipWriter.Close()
	subtitleBookPreviewZipFile.Close()

	// Create WEBP cover file
	coverWEBPFilePath := filepath.Join(subtitleFolderPath, "cover.webp")
	cmd := exec.Command("cwebp", coverFilePath, "-resize", "650", "0", "-q", "100", "-o", coverWEBPFilePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("sSubtitle - failed to resize and create WEBP cover for subtitle: %w", err)
	}
	os.Remove(coverFilePath)

	return nil
}

func addFileToZip(zipWriter *zip.Writer, file *zip.File) error {
	pageFileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer pageFileReader.Close()

	info := file.FileInfo()
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = file.Name
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, pageFileReader)
	if err != nil {
		return err
	}

	return nil
}
