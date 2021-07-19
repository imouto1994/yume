package route

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	httpServer "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/model"
	"github.com/imouto1994/yume/internal/service"
	"go.uber.org/zap"
)

type HandlerLibrary struct {
	db             sqlite.DB
	serviceLibrary service.ServiceLibrary
	validate       *validator.Validate
}

func NewHandlerLibrary(db sqlite.DB, v *validator.Validate, s service.ServiceLibrary) *HandlerLibrary {
	return &HandlerLibrary{
		db:             db,
		serviceLibrary: s,
		validate:       v,
	}
}

func (h *HandlerLibrary) InitializeRoutes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.handleCreateLibrary())
	r.Get("/", h.handleGetLibraries())
	r.Delete("/{libraryID}", h.handleDeleteLibrary())
	r.Post("/{libraryID}/scan", h.handleScanLibrary())

	return r
}

func (h *HandlerLibrary) handleCreateLibrary() http.HandlerFunc {
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
			httpServer.RespondBadRequestError(w, "request body is not in JSON format", fmt.Errorf("hLibrary - request body is not in JSON format for creating library: %w ", err))
			return
		}

		err = h.validate.Struct(body)
		if err != nil {
			httpServer.RespondBadRequestError(w, "request body is invalid", fmt.Errorf("hLibrary - request JSON body is not valid for creating library: %w ", err))
			return
		}

		newLibrary := &model.Library{
			Name: body.Name,
			Root: body.Root,
		}

		err = h.serviceLibrary.CreateLibrary(ctx, h.db, newLibrary)
		if err != nil {
			httpServer.RespondError(w, "failed to create library", fmt.Errorf("hLibrary - failed to use service Library to create library: %w", err))
			return
		}

		resp := response{
			Library: newLibrary,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerLibrary) handleGetLibraries() http.HandlerFunc {
	type response struct {
		Libraries []*model.Library `json:"libraries"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		libraries, err := h.serviceLibrary.GetLibraries(ctx, h.db)
		if err != nil {
			httpServer.RespondError(w, "failed to get libraries", fmt.Errorf("hLibrary - failed to user service Library to get libraries: %w", err))
			return
		}

		resp := response{
			Libraries: libraries,
		}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerLibrary) handleDeleteLibrary() http.HandlerFunc {
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		libraryID := chi.URLParam(r, "libraryID")

		tx, err := h.db.BeginTxx(ctx, nil)
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to delete library", fmt.Errorf("hLibrary - failed to begin SQL transaction for deleting library: %w", err))
			return
		}

		err = h.serviceLibrary.DeleteLibraryByID(ctx, tx, libraryID)
		if err != nil {
			tx.Rollback()
			httpServer.RespondError(w, "failed to delete library", fmt.Errorf("hLibrary - failed to user service Library to delete library: %w", err))
			return
		}

		err = tx.Commit()
		if err != nil {
			httpServer.RespondInternalServerError(w, "failed to delete library", fmt.Errorf("hLibrary - failed to commit SQL transaction for deleting library: %w", err))
			return
		}

		resp := response{}
		httpServer.RespondJSON(w, 200, resp)
	}
}

func (h *HandlerLibrary) handleScanLibrary() http.HandlerFunc {
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		libraryID := chi.URLParam(r, "libraryID")

		library, err := h.serviceLibrary.GetLibraryByID(ctx, h.db, libraryID)
		if err != nil {
			httpServer.RespondError(w, "failed to get library", err)
			return
		}

		go func() {
			start := time.Now()
			ctx := context.Background()
			tx, err := h.db.BeginTxx(ctx, nil)
			if err != nil {
				zap.L().Error("hLibrary - failed to begin SQL transaction for scanning library", zap.Error(err))
				return
			}
			err = h.serviceLibrary.ScanLibrary(ctx, tx, library)
			if err != nil {
				zap.L().Error("hLibrary - failed to scan and update library", zap.Error(err))
				tx.Rollback()
				return
			}
			err = tx.Commit()
			if err != nil {
				zap.L().Error("hLibrary - failed to commit SQL transaction for scanning library", zap.Error(err))
			}
			elapsed := time.Since(start)
			zap.L().Info("hLibrary - successfully scanned and updated library", zap.Duration("duration", elapsed))
		}()

		resp := response{}
		httpServer.RespondJSON(w, 200, resp)
	}
}
