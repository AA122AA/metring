package handler

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"reflect"
	"strings"

	"github.com/AA122AA/metring/internal/server/constants"
	"github.com/AA122AA/metring/internal/server/domain"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Metrics interface {
	Parse(mType string, mName string, value string, handler string) (*domain.MetricsJSON, error)
	Update(ctx context.Context, metric *domain.MetricsJSON) error
	Updates(ctx context.Context, metrics []*domain.MetricsJSON) error
	Get(ctx context.Context, metric *domain.MetricsJSON) (string, error)
	GetJSON(ctx context.Context, metric *domain.MetricsJSON) (*domain.MetricsJSON, error)
	GetAll(ctx context.Context) (map[string]*domain.Metrics, error)
}

type Saver interface {
	WriteSync(data *domain.MetricsJSON) error
	WriteSyncBatch(data []*domain.MetricsJSON) error
}

type MetricsHandler struct {
	srv      Metrics
	saver    Saver
	lg       *zap.Logger
	tmplPath string
}

func NewMetricsHandler(ctx context.Context, tPath string, srv Metrics, saver Saver) *MetricsHandler {
	h := &MetricsHandler{
		srv:      srv,
		lg:       zctx.From(ctx).Named("metrics handler"),
		tmplPath: tPath,
	}

	v := reflect.ValueOf(saver)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		saver = nil
	}
	h.saver = saver

	return h
}

func (h MetricsHandler) All(w http.ResponseWriter, r *http.Request) {
	templates, err := template.ParseGlob(h.tmplPath)
	if err != nil {
		h.lg.Error("no templates within path", zap.String("path", h.tmplPath), zap.Error(err))
		http.Error(w, "no html templates", http.StatusInternalServerError)
		return
	}

	metrics, err := h.srv.GetAll(r.Context())
	if err != nil {
		h.lg.Error("service returned no data", zap.Error(err))
		http.Error(w, "no data", http.StatusNotFound)
		return
	}
	data := struct {
		Metrics map[string]*domain.Metrics
	}{
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	templates.ExecuteTemplate(w, "metrics.html", data)
}

func (h MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	mName := chi.URLParam(r, "mName")
	mType := chi.URLParam(r, "mType")
	// mName := r.PathValue("mName")
	// mType := r.PathValue("mType")

	data, err := h.srv.Parse(mType, mName, "", constants.Get)
	if err != nil {
		http.Error(w, "smth went wrong", http.StatusInternalServerError)
		h.lg.Error("what happend", zap.String("name", mName), zap.Error(err))
		return
	}

	m, err := h.srv.Get(r.Context(), data)
	if err != nil {
		// if err.Error() == "err from repo: data not found" {
		var er *repository.EmptyRepoError
		if errors.Is(err, er) {
			http.Error(w, "No metric with this name", http.StatusNotFound)
			h.lg.Error("no metric with provided name", zap.String("name", mName), zap.Error(err))
			return
		}
		if strings.Contains(err.Error(), "has different types between data and repo") {
			http.Error(w, "No metric with this type", http.StatusNotFound)
			h.lg.Error("no metric with provided type", zap.String("type", mType), zap.Error(err))
			return
		}
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("got error in repo", zap.Error(err))
		return
	}

	h.lg.Debug("gonna give metric with name", zap.String("name", mName))

	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(m))
}

func (h MetricsHandler) GetJSON(w http.ResponseWriter, r *http.Request) {
	data := domain.MetricsJSON{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("error while decoding", zap.Error(err))
		return
	}

	m, err := h.srv.GetJSON(r.Context(), &data)
	if err != nil {
		var er *repository.EmptyRepoError
		if errors.Is(err, er) {
			// if err.Error() == "err from repo: data not found" {
			http.Error(w, "No metric with this name", http.StatusNotFound)
			h.lg.Error("no metric with provided name", zap.String("name", data.ID), zap.Error(err))
			return
		}
		if strings.Contains(err.Error(), "has different types between data and repo") {
			http.Error(w, "No metric with this type", http.StatusNotFound)
			h.lg.Error("no metric with provided type", zap.String("type", data.MType), zap.Error(err))
			return
		}
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("got error in repo", zap.Error(err))
		return
	}

	h.lg.Debug("gonna give metric with name", zap.String("name", data.ID))
	res, err := json.Marshal(m)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("got error in repo", zap.Error(err))
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
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

	data, err := h.srv.Parse(mType, mName, value, constants.Update)
	if err != nil {
		h.lg.Error("metrics type or value is incorrect", zap.Error(err))
		http.Error(w, "тип или значение некорректно", http.StatusBadRequest)
		return
	}

	err = h.srv.Update(r.Context(), data)
	if err != nil {
		h.lg.Error("error while updating metric", zap.Error(err))
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if h.saver != nil {
		h.lg.Info("saver is not nil", zap.Any("h.saver", h.saver))
		err = h.saver.WriteSync(data)
		if err != nil {
			h.lg.Error("error while writing to file", zap.Error(err))
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	h.lg.Debug("got new metric", zap.String("name", mName), zap.String("value", value))

	w.WriteHeader(http.StatusOK)
}

func (h MetricsHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	metric := domain.MetricsJSON{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&metric)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("error while decoding", zap.Error(err))
		return
	}

	err = h.srv.Update(r.Context(), &metric)
	if err != nil {
		h.lg.Error("metrics type or value is incorrect")
		http.Error(w, "тип или значение некорректно", http.StatusBadRequest)
		return
	}

	if h.saver != nil {
		err = h.saver.WriteSync(&metric)
		if err != nil {
			h.lg.Error("error while writing to file", zap.Error(err))
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h MetricsHandler) Updates(w http.ResponseWriter, r *http.Request) {
	metrics := make([]*domain.MetricsJSON, 0, 20)
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		h.lg.Error("error while decoding", zap.Error(err))
		return
	}

	err = h.srv.Updates(r.Context(), metrics)
	if err != nil {
		h.lg.Error("metrics type or value is incorrect")
		http.Error(w, "тип или значение некорректно", http.StatusBadRequest)
		return
	}

	if h.saver != nil {
		err = h.saver.WriteSyncBatch(metrics)
		if err != nil {
			h.lg.Error("error while writing to file", zap.Error(err))
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
