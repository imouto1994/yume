package route

import (
	"net/http"

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

	r.Get("/{titleID}/cover", h.handleGetTitleCoverFile())
	r.Get("/{titleID}/books", h.handleGetTitleBooks())

	return r
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
			httpServer.RespondInternalServerError(w, "failed to get books for specitic title ID", err)
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
			httpServer.RespondInternalServerError(w, "failed to stream title cover", err)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(200)
	}
}
