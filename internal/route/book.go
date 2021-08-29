package route

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/service"
)

type HandlerBook struct {
	db          sqlite.DB
	serviceBook service.ServiceBook
}

func NewHandlerBook(db sqlite.DB, sBook service.ServiceBook) *HandlerBook {
	return &HandlerBook{
		db:          db,
		serviceBook: sBook,
	}
}

func (h *HandlerBook) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/{bookID}", h.handleGetBookByID())
	r.Get("/{bookID}/pages", h.handleGetBookPages())
	r.Get("/{bookID}/page/{pageIndex}", h.handleGetBookPageFile())
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
