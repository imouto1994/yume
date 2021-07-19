package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator"

	"github.com/imouto1994/yume/internal/infra/config"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/repository"
	"github.com/imouto1994/yume/internal/service"
)

func CreateRouter(cfg *config.Config, db sqlite.DB, v *validator.Validate) http.Handler {
	// Initialize repositories
	repositoryLibrary := repository.NewRepositoryLibrary()
	repositoryTitle := repository.NewRepositoryTitle()
	repositoryBook := repository.NewRepositoryBook()
	repositoryPage := repository.NewRepositoryPage()

	// Initialize services
	serviceImage := service.NewServiceImage()
	serviceArchive := service.NewServiceArchive()
	serviceScanner := service.NewServiceScanner(serviceImage, serviceArchive)
	serviceBook := service.NewServiceBook(repositoryBook, repositoryPage, serviceArchive, serviceImage)
	serviceTitle := service.NewServiceTitle(repositoryTitle, serviceBook)
	serviceLibrary := service.NewServiceLibrary(repositoryLibrary, serviceScanner, serviceTitle, serviceBook)

	// Initialize handlers
	handlerLibrary := NewHandlerLibrary(db, v, serviceLibrary)
	hanlderBook := NewHandlerBook(db, serviceBook)
	handlerTitle := NewHandlerTitle(db, serviceTitle, serviceBook)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	r.Mount("/api/library", handlerLibrary.InitializeRoutes())
	r.Mount("/api/title", handlerTitle.InitializeRoutes())
	r.Mount("/api/book", hanlderBook.InitializeRoutes())

	return r
}
