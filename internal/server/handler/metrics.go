package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	models "github.com/AA122AA/metring/internal/server/model"
)

type Metrics interface {
	Update(string, string, string) error
	Get(string) (*models.Metrics, error)
}

type MetricsHandler struct {
	srv Metrics
}

func NewMetricsHandler(srv Metrics) *MetricsHandler {
	return &MetricsHandler{srv: srv}
}

func (h MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	mName := r.PathValue("mName")
	m, err := h.srv.Get(mName)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		log.Printf("got error - %v", err)
		return
	}

	jsondata, err := json.Marshal(m)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		log.Printf("got error - %v", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsondata)
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
