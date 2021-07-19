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
	r.Get("/{bookID}/page/{pageNumber}", h.handleGetBookPageFile())

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

func (h *HandlerBook) handleGetBookPageFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")
		pageNumberString := chi.URLParam(r, "pageNumber")
		pageNumber, err := strconv.Atoi(pageNumberString)
		if err != nil {
			httpServer.RespondBadRequestError(w, "page number is invalid", fmt.Errorf("hBook - page number param is not a number for streaming book page: %w", err))
			return
		}

		w.Header().Set("Cache-Control", "max-age=86400,public")
		_, err = h.serviceBook.StreamBookPageByID(ctx, h.db, w, bookID, pageNumber)
		if err != nil {
			httpServer.RespondError(w, "failed to stream book page", fmt.Errorf("hBook - failed to use service Book to stream book page: %w", err))
			return
		}

		w.WriteHeader(200)
	}
}
