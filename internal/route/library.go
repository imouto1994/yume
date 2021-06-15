package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/imouto1994/yume/internal/service"
)

type LibraryHandler struct {
	libraryService *service.LibraryService
}

func NewLibraryHandler(s *service.LibraryService) *LibraryHandler {
	return &LibraryHandler{
		libraryService: s,
	}
}

func (h *LibraryHandler) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.createLibrary)
	r.Get("/", h.getLibraries)

	return r
}

func (h *LibraryHandler) createLibrary(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create Library"))
}

func (h *LibraryHandler) getLibraries(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get Libraries"))
}
