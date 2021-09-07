package route

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/service"
	"go.uber.org/zap"
)

type HandlerBook struct {
	db          sqlite.DB
	serviceBook service.ServiceBook
	validate    *validator.Validate
}

func NewHandlerBook(db sqlite.DB, sBook service.ServiceBook, v *validator.Validate) *HandlerBook {
	return &HandlerBook{
		db:          db,
		serviceBook: sBook,
		validate:    v,
	}
}

func (h *HandlerBook) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/{bookID}", h.handleGetBookByID())
	r.Get("/{bookID}/pages", h.handleGetBookPages())
	r.Get("/{bookID}/page/{pageIndex}", h.handleGetBookPageFile())
	r.Put("/{bookID}/page/{pageNumber}/favorite", h.handleUpdateBookPageFavorite())
	r.Get("/{bookID}/previews", h.handleGetBookPreviews())
	r.Get("/{bookID}/preview/{previewIndex}", h.handleGetBookPreviewFile())

	return r
}

func (h *HandlerBook) handleGetBookByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")

		book, err := h.serviceBook.GetBookByID(ctx, h.db, bookID)
		if err != nil {
			httpServer.RespondError(w, "failed to get book", fmt.Errorf("hBook - failed to use service Book to get book by book ID: %w", err))
			return
		}

		httpServer.RespondJSON(w, 200, book)
	}
}

func (h *HandlerBook) handleGetBookPages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")

		pages, err := h.serviceBook.GetBookPages(ctx, h.db, bookID)
		if err != nil {
			httpServer.RespondError(w, "failed to get book pages", fmt.Errorf("hBook - failed to use service Book to get book pages: %w", err))
			return
		}

		httpServer.RespondJSON(w, 200, pages)
	}
}

func (h *HandlerBook) handleGetBookPreviews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")

		previews, err := h.serviceBook.GetBookPreviews(ctx, h.db, bookID)
		if err != nil {
			httpServer.RespondError(w, "failed to get book previews", fmt.Errorf("hBook - failed to use service Book to get book previews: %w", err))
			return
		}

		httpServer.RespondJSON(w, 200, previews)
	}
}

func (h *HandlerBook) handleGetBookPageFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")
		pageIndexString := chi.URLParam(r, "pageIndex")
		pageIndex, err := strconv.Atoi(pageIndexString)
		if err != nil {
			httpServer.RespondBadRequestError(w, "page index is invalid", fmt.Errorf("hBook - page index param is not a number for streaming book page: %w", err))
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		_, err = h.serviceBook.StreamBookPageByID(ctx, h.db, w, bookID, pageIndex)
		if err != nil {
			httpServer.RespondError(w, "failed to stream book page", fmt.Errorf("hBook - failed to use service Book to stream book page: %w", err))
			return
		}

		w.WriteHeader(200)
	}
}

func (h *HandlerBook) handleGetBookPreviewFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")
		previewIndexString := chi.URLParam(r, "previewIndex")
		previewIndex, err := strconv.Atoi(previewIndexString)
		if err != nil {
			httpServer.RespondBadRequestError(w, "preview index is invalid", fmt.Errorf("hBook - preview index param is not a number for streaming book preview: %w", err))
			return
		}

		w.Header().Set("Cache-Control", "max-age=86400,public")
		_, err = h.serviceBook.StreamBookPreviewByID(ctx, h.db, w, bookID, previewIndex)
		if err != nil {
			httpServer.RespondError(w, "failed to stream book preview", fmt.Errorf("hBook - failed to use service Book to stream book preview: %w", err))
			return
		}

		w.WriteHeader(200)
	}
}

func (h *HandlerBook) handleUpdateBookPageFavorite() http.HandlerFunc {
	type request struct {
		Favorite *int `json:"favorite" validate:"required"`
	}

	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse params
		bookID := chi.URLParam(r, "bookID")
		pageNumberString := chi.URLParam(r, "pageNumber")
		pageNumber, err := strconv.Atoi(pageNumberString)
		if err != nil {
			httpServer.RespondBadRequestError(w, "page number is invalid", fmt.Errorf("hBook - page number param is not a number for updating book page favorite: %w", err))
			return
		}

		// Parse body
		var body request
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			httpServer.RespondBadRequestError(w, "request body is not in JSON format", fmt.Errorf("hBook - request body is not in JSON format for updating book page favorite: %w ", err))
			return
		}

		err = h.validate.Struct(body)
		if err != nil {
			httpServer.RespondBadRequestError(w, "request body is invalid", fmt.Errorf("hBook - request JSON body is not valid for updating book page favorite: %w ", err))
			return
		}

		err = h.serviceBook.UpdateBookPageFavorite(ctx, h.db, bookID, pageNumber, *body.Favorite)
		if err != nil {
			httpServer.RespondError(w, "failed to update book page favorite", fmt.Errorf("hBook - failed to use service Book to update book page favorite: %w", err))
			return
		}

		go func() {
			ctx := context.Background()

			book, err := h.serviceBook.GetBookByID(ctx, h.db, bookID)
			if err != nil {
				zap.L().Error("hBook - failed to get book data to save book metadata", zap.Error(err))
				return
			}

			pages, err := h.serviceBook.GetBookPages(ctx, h.db, bookID)
			if err != nil {
				zap.L().Error("hBook - failed to get updated pages data to save book metadata", zap.Error(err))
				return
			}

			favoriteIndices := []int{}
			for index, page := range pages {
				if page.Favorite == 1 {
					favoriteIndices = append(favoriteIndices, index)
				}
			}

			titleFolderPath := filepath.Dir(book.URL)
			metadataFilePath := filepath.Join(titleFolderPath, fmt.Sprintf("%s.json", book.Name))

			file, _ := json.MarshalIndent(favoriteIndices, "", " ")
			ioutil.WriteFile(metadataFilePath, file, os.ModePerm)
		}()

		resp := response{}
		httpServer.RespondJSON(w, 200, resp)
	}
}
