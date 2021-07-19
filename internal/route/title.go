package route

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/service"
)

type HandlerTitle struct {
	db           sqlite.DB
	serviceTitle service.ServiceTitle
	serviceBook  service.ServiceBook
}

func NewHandlerTitle(db sqlite.DB, sTitle service.ServiceTitle, sBook service.ServiceBook) *HandlerTitle {
	return &HandlerTitle{
		db:           db,
		serviceTitle: sTitle,
		serviceBook:  sBook,
	}
}

func (h *HandlerTitle) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.handleGetTitles())
	r.Get("/count", h.handleCountGetTitles())
	r.Get("/{titleID}", h.handleGetTitleByID())
	r.Get("/{titleID}/cover", h.handleGetTitleCoverFile())
	r.Get("/{titleID}/books", h.handleGetTitleBooks())

	return r
}

func (h *HandlerTitle) handleGetTitles() http.HandlerFunc {
	type response struct {
		Titles []*model.Title `json:"titles"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		queryValues := r.URL.Query()

		var err error

		pageString := queryValues.Get("page")
		pageNumber := 0
		if pageString != "" {
			pageNumber, err = strconv.Atoi(pageString)
			if err != nil {
				httpServer.RespondBadRequestError(w, "page number is invalid", fmt.Errorf("hTitle - page number query is not number for searching titles: %w", err))
				return
			}
		}

		sizeString := queryValues.Get("size")
		sizeNumber := 24
		if sizeString != "" {
			sizeNumber, err = strconv.Atoi(sizeString)
			if err != nil {
				httpServer.RespondBadRequestError(w, "size number is invalid", fmt.Errorf("hTitle - size number query is not number for searching titles: %w", err))
				return
			}
		}

		titleQuery := &model.TitleQuery{
			LibraryIDs: queryValues["library_id"],
			Page:       pageNumber,
			Size:       sizeNumber,
			Sort:       queryValues.Get("sort"),
			Search:     queryValues.Get("search"),
		}

		titles, err := h.serviceTitle.SearchTitles(ctx, h.db, titleQuery)
		if err != nil {
			httpServer.RespondError(w, "failed to search for titles with given query", fmt.Errorf("hTitle - failed to use service Title to search titles: %w", err))
			return
		}

		resp := response{
			Titles: titles,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerTitle) handleCountGetTitles() http.HandlerFunc {
	type response struct {
		Count int `json:"count"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queryValues := r.URL.Query()

		titleQuery := &model.TitleQuery{
			LibraryIDs: queryValues["library_id"],
			Search:     queryValues.Get("search"),
		}

		count, err := h.serviceTitle.CountSearchTitles(ctx, h.db, titleQuery)
		if err != nil {
			httpServer.RespondError(w, "failed to count number of search results with given query", fmt.Errorf("hTitle - failed to use service Title to count for total search results: %w", err))
			return
		}

		resp := response{
			Count: count,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerTitle) handleGetTitleByID() http.HandlerFunc {
	type response struct {
		Title *model.Title `json:"title"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		titleID := chi.URLParam(r, "titleID")

		title, err := h.serviceTitle.GetTitleByID(ctx, h.db, titleID)
		if err != nil {
			httpServer.RespondError(w, "failed to get title", fmt.Errorf("hTitle - failed to use service Title to get title by title ID: %w", err))
			return
		}

		resp := response{
			Title: title,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerTitle) handleGetTitleBooks() http.HandlerFunc {
	type response struct {
		Books []*model.Book `json:"books"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		titleID := chi.URLParam(r, "titleID")

		books, err := h.serviceBook.GetBooksByTitleID(ctx, h.db, titleID)
		if err != nil {
			httpServer.RespondError(w, "failed to get books for specitic title ID", fmt.Errorf("hTitle - failed to use service Book to get books by title ID: %w", err))
			return
		}

		resp := response{
			Books: books,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerTitle) handleGetTitleCoverFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		titleID := chi.URLParam(r, "titleID")

		err := h.serviceTitle.StreamTitleCoverByID(ctx, h.db, w, titleID)
		if err != nil {
			httpServer.RespondError(w, "failed to stream title cover", fmt.Errorf("hTitle - failed to use service Title to stream title cover: %w", err))
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(200)
	}
}
