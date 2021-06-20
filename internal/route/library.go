package route

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/service"
)

type LibraryHandler struct {
	libraryService libraryService
	validate       validate
}

type libraryService interface {
	CreateLibrary(context.Context, *model.Library) error
	GetLibraries(context.Context) ([]*model.Library, error)
	DeleteLibrary(context.Context, string) error
}

type validate interface {
	Struct(s interface{}) error
}

func NewLibraryHandler(s *service.LibraryService, v validate) *LibraryHandler {
	return &LibraryHandler{
		libraryService: s,
		validate:       v,
	}
}

func (h *LibraryHandler) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.handleCreateLibrary())
	r.Get("/", h.handleGetLibraries())
	r.Delete("/${libraryID}", h.handleDeleteLibrary())

	return r
}

func (h *LibraryHandler) handleCreateLibrary() http.HandlerFunc {
	type request struct {
		Name string `json:"name" validate:"required"`
		Root string `json:"root" validate:"required"`
	}

	type response struct {
		Library *model.Library `json:"library"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var body request
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			httpServer.RespondBadRequest(w, "request body is not in JSON format", err)
			return
		}

		err = h.validate.Struct(body)
		if err != nil {
			httpServer.RespondBadRequest(w, "request body is invalid", err)
			return
		}

		newLibrary := &model.Library{
			Name: body.Name,
			Root: body.Root,
		}

		err = h.libraryService.CreateLibrary(ctx, newLibrary)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to create library", err)
			return
		}

		resp := response{
			Library: newLibrary,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *LibraryHandler) handleGetLibraries() http.HandlerFunc {
	type response struct {
		Libraries []*model.Library `json:"libraries"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		libraries, err := h.libraryService.GetLibraries(ctx)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to get libraries", err)
			return
		}

		resp := response{
			Libraries: libraries,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *LibraryHandler) handleDeleteLibrary() http.HandlerFunc {
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		libraryID := chi.URLParam(r, "libraryID")

		err := h.libraryService.DeleteLibrary(ctx, libraryID)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to delete library", err)
			return
		}

		resp := response{}
		httpServer.RespondJSON(w, 200, resp)
	}
}
