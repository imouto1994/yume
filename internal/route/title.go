package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/service"
)

type HandlerTitle struct {
	db              sqlite.DB
	serviceTitle    service.ServiceTitle
	serviceBook     service.ServiceBook
	serviceSubtitle service.ServiceSubtitle
	validate        *validator.Validate
}

func NewHandlerTitle(db sqlite.DB, sTitle service.ServiceTitle, sBook service.ServiceBook, sSubtitle service.ServiceSubtitle, v *validator.Validate) *HandlerTitle {
	return &HandlerTitle{
		db:              db,
		serviceTitle:    sTitle,
		serviceBook:     sBook,
		serviceSubtitle: sSubtitle,
		validate:        v,
	}
}

func (h *HandlerTitle) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.handleGetTitles())
	r.Get("/count", h.handleCountGetTitles())
	r.Get("/{titleID}", h.handleGetTitleByID())
	r.Get("/{titleID}/cover", h.handleGetTitleCoverFile())
	r.Get("/{titleID}/books", h.handleGetTitleBooks())
	r.Post("/subtitle", h.handleCreateSubtitle())

	return r
}

func (h *HandlerTitle) handleGetTitles() http.HandlerFunc {
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

		httpServer.RespondJSON(w, 200, titles)
	}
}

func (h *HandlerTitle) handleCountGetTitles() http.HandlerFunc {
	type response struct {
		Value int `json:"value"`
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
			Value: count,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerTitle) handleGetTitleByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		titleID := chi.URLParam(r, "titleID")

		title, err := h.serviceTitle.GetTitleByID(ctx, h.db, titleID)
		if err != nil {
			httpServer.RespondError(w, "failed to get title", fmt.Errorf("hTitle - failed to use service Title to get title by title ID: %w", err))
			return
		}

		httpServer.RespondJSON(w, 200, title)
	}
}

func (h *HandlerTitle) handleGetTitleBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		titleID := chi.URLParam(r, "titleID")

		books, err := h.serviceBook.GetBooksByTitleID(ctx, h.db, titleID)
		if err != nil {
			httpServer.RespondError(w, "failed to get books for specitic title ID", fmt.Errorf("hTitle - failed to use service Book to get books by title ID: %w", err))
			return
		}

		httpServer.RespondJSON(w, 200, books)
	}
}

func (h *HandlerTitle) handleGetTitleCoverFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		titleID := chi.URLParam(r, "titleID")

		w.Header().Set("Cache-Control", "max-age=86400,public")
		err := h.serviceTitle.StreamTitleCoverByID(ctx, h.db, w, titleID)
		if err != nil {
			httpServer.RespondError(w, "failed to stream title cover", fmt.Errorf("hTitle - failed to use service Title to stream title cover: %w", err))
			return
		}

		w.WriteHeader(200)
	}
}

func (h *HandlerTitle) handleCreateSubtitle() http.HandlerFunc {
	type request struct {
		Name            string `json:"name" validate:"required"`
		Author          string `json:"author" validate:"required"`
		PageStartNumber *int   `json:"page_start_number" validate:"required"`
		PageEndNumber   *int   `json:"page_end_number" validate:"required"`
		BookID          string `json:"book_id" validate:"required"`
		LibraryID       string `json:"library_id" validate:"required"`
	}

	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var body request
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			httpServer.RespondBadRequestError(w, "request body is not in JSON format", fmt.Errorf("hTitle - request body is not in JSON format for creating subtitle: %w ", err))
			return
		}

		err = h.validate.Struct(body)
		if err != nil {
			httpServer.RespondBadRequestError(w, "request body is invalid", fmt.Errorf("hTitle - request JSON body is not valid for creating subtitle: %w ", err))
			return
		}

		newSubtitle := &model.Subtitle{
			Name:            body.Name,
			Author:          body.Author,
			PageStartNumber: *body.PageStartNumber,
			PageEndNumber:   *body.PageEndNumber,
			BookID:          body.BookID,
			LibraryID:       body.LibraryID,
		}

		err = h.serviceSubtitle.CreateSubtitle(ctx, h.db, newSubtitle)
		if err != nil {
			httpServer.RespondError(w, "failed to create library", fmt.Errorf("hTitle - failed to use service Subtitle to create subtitle: %w", err))
			return
		}

		resp := response{}
		httpServer.RespondJSON(w, 200, resp)
	}
}
