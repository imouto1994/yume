package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator"

	"github.com/imouto1994/yume/internal/infra/config"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/repository"
	"github.com/imouto1994/yume/internal/service"
)

func CreateRouter(cfg *config.Config, db sqlite.DB, v *validator.Validate) http.Handler {
	// Initialize repositories
	libraryRepository := repository.NewLibraryRepository()

	// Initialize services
	libraryService := service.NewLibraryService(libraryRepository, db)

	// Initialize handlers
	libraryHandler := NewLibraryHandler(libraryService, v)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/api/library", libraryHandler.InitializeRoutes())

	return r
}
