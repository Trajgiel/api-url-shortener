package delete

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	resp "github.com/Trajgiel/api-url-shortener/internal/lib/api/response"
	"github.com/Trajgiel/api-url-shortener/internal/lib/logger/sl"
	"github.com/Trajgiel/api-url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLDeleter interface {
	DeleteURL(id int64) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			log.Info("invalid id", "id", idParam)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		err = urlDeleter.DeleteURL(id)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.Int64("id", id))
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("url deleted", slog.Int64("id", id))

		render.JSON(w, r, resp.OK())
	}
}
