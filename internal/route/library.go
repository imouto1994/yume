package route

import (
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
	CreateLibrary(*model.Library) error
	GetLibraries() ([]*model.Library, error)
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
		var body request

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			httpServer.RespondBadRequest(w, "Request body is not in JSON format", err)
			return
		}

		err = h.validate.Struct(body)
		if err != nil {
			httpServer.RespondBadRequest(w, "Request body is invalid", err)
			return
		}

		newLibrary := &model.Library{
			Name: body.Name,
			Root: body.Root,
		}

		err = h.libraryService.CreateLibrary(newLibrary)
		if err != nil {
			httpServer.RespondInternalServerError(w, "Failed to create library", err)
			return
		}

		resp := response{
			Library: newLibrary,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *LibraryHandler) handleGetLibraries() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Get Libraries"))
	}
}
