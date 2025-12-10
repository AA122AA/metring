package handler

import (
	"context"
	"html/template"
	"net/http"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metrics interface {
	Update(mName string, mType string, value string) error
	Get(mType string, mName string) (string, error)
	GetAll() (map[string]*models.Metrics, error)
}

type MetricsHandler struct {
	srv      Metrics
	lg       *zap.Logger
	tmplPath string
}

func NewMetricsHandler(ctx context.Context, tPath string, srv Metrics) *MetricsHandler {
	return &MetricsHandler{
		srv:      srv,
		lg:       zctx.From(ctx).Named("metrics handler"),
		tmplPath: tPath,
	}
}

func (h MetricsHandler) All(w http.ResponseWriter, r *http.Request) {
	templates, err := template.ParseGlob(h.tmplPath)
	if err != nil {
		h.lg.Error("no templates within path", zap.String("path", h.tmplPath), zap.Error(err))
		http.Error(w, "no html templates", http.StatusInternalServerError)
		return
	}

	metrics, err := h.srv.GetAll()
	if err != nil {
		h.lg.Error("service returned no data", zap.Error(err))
		http.Error(w, "no data", http.StatusNotFound)
		return
	}
	data := struct {
		Metrics map[string]*models.Metrics
	}{
		Metrics: metrics,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "text/html")
	templates.ExecuteTemplate(w, "metrics.html", data)
}

func (h MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	mName := chi.URLParam(r, "mName")
	mType := chi.URLParam(r, "mType")
	// mName := r.PathValue("mName")
	// mType := r.PathValue("mType")

	m, err := h.srv.Get(mType, mName)
	if err != nil {
		if err.Error() == "err from repo: data not found" {
			http.Error(w, "No metric with this name", http.StatusNotFound)
			h.lg.Error("no metric with provided name", zap.String("name", mName), zap.Error(err))
			return
		}
		if err.Error() == "wrong metric type" {
			http.Error(w, "No metric with this type", http.StatusNotFound)
			h.lg.Error("no metric with provided type", zap.String("type", mType), zap.Error(err))
			return
		}
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("got error in repo", zap.Error(err))
		return
	}

	h.lg.Debug("gonna give metric with name", zap.String("name", mName))

	w.Header().Set("Content-type", "http/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(m))
}

func (h MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	mType := r.PathValue("mType")
	mName := r.PathValue("mName")
	value := r.PathValue("value")
	if mName == "" {
		h.lg.Error("no metrics name was provided")
		http.Error(w, "имя метрики не указано", http.StatusNotFound)
		return
	}

	err := h.srv.Update(mName, mType, value)
	if err != nil {
		h.lg.Error("metrics type or value is incorrect")
		http.Error(w, "тип или значение некорректно", http.StatusBadRequest)
		return
	}

	h.lg.Debug("got new metric", zap.String("name", mName), zap.String("value", value))

	w.WriteHeader(http.StatusOK)
}
