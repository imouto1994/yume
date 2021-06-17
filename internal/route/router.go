package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator"

	"github.com/imouto1994/yume/internal/infra/config"
	"github.com/imouto1994/yume/internal/repository"
	"github.com/imouto1994/yume/internal/service"
	"github.com/jmoiron/sqlx"
)

func CreateRouter(cfg *config.Config, db *sqlx.DB, v *validator.Validate) http.Handler {
	// Initialize repositories
	libraryRepository := repository.NewLibraryRepository(db)

	// Initialize services
	libraryService := service.NewLibraryService(libraryRepository)

	// Initialize handlers
	libraryHandler := NewLibraryHandler(libraryService, v)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/api/library", libraryHandler.InitializeRoutes())

	return r
}
