package route

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
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
	type response struct {
		Book *model.Book `json:"book"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")

		book, err := h.serviceBook.GetBookByID(ctx, h.db, bookID)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to get book", err)
			return
		}

		resp := response{
			Book: book,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerBook) handleGetBookPages() http.HandlerFunc {
	type response struct {
		Pages []*model.Page `json:"pages"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")

		pages, err := h.serviceBook.GetBookPages(ctx, h.db, bookID)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to get book pages", err)
			return
		}

		resp := response{
			Pages: pages,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerBook) handleGetBookPageFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookID := chi.URLParam(r, "bookID")
		pageNumberString := chi.URLParam(r, "pageNumber")
		pageNumber, err := strconv.Atoi(pageNumberString)
		if err != nil {
			httpServer.RespondBadRequest(w, "page number is invalid", err)
		}

		err = h.serviceBook.StreamBookPageByID(ctx, h.db, w, bookID, pageNumber)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to stream book page", err)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(200)
	}
}
