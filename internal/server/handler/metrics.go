package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	models "github.com/AA122AA/metring/internal/server/model"
	"github.com/go-chi/chi/v5"
)

type Metrics interface {
	Update(string, string, string) error
	Get(string, string) (string, error)
	GetAll() (map[string]*models.Metrics, error)
}

type MetricsHandler struct {
	srv Metrics
}

func NewMetricsHandler(srv Metrics) *MetricsHandler {
	return &MetricsHandler{srv: srv}
}

func (h MetricsHandler) All(w http.ResponseWriter, r *http.Request) {
	templates, err := template.ParseGlob("/home/artem/Documents/development/Yandex.Practicum/metring/internal/server/templates/*.html")
	if err != nil {
		http.Error(w, "no html templates", http.StatusNotFound)
		return
	}

	metrics, err := h.srv.GetAll()
	if err != nil {
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
			log.Printf("got error - %v", err)
			return
		}
		if err.Error() == "wrong metric type" {
			http.Error(w, "No metric with this type", http.StatusNotFound)
			log.Printf("got error - %v", err)
			return
		}
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		log.Printf("got error in repo - %v", err)
		return
	}

	w.Header().Set("Content-type", "http/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(m))
}

func (h MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	mType := r.PathValue("mType")
	mName := r.PathValue("mName")
	value := r.PathValue("value")
	if mName == "" {
		http.Error(w, "имя метрики не указано", http.StatusNotFound)
		return
	}

	fmt.Printf("got new metric - %v, value - %v\n", mName, value)

	err := h.srv.Update(mName, mType, value)
	if err != nil {
		http.Error(w, "тип или значение некорректно", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
