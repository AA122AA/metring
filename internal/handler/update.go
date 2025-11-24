package handler

import (
	"encoding/json"
	"log"
	"net/http"

	models "github.com/AA122AA/metring/internal/model"
)

type Metrics interface {
	Update(string, string, string) error
	Get(string) (*models.Metrics, error)
}

type Handler struct {
	srv Metrics
}

func NewHandler(srv Metrics) *Handler {
	return &Handler{srv: srv}
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
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

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	mType := r.PathValue("mType")
	mName := r.PathValue("mName")
	value := r.PathValue("value")
	if mName == "" {
		http.Error(w, "имя метрики не указано", http.StatusNotFound)
		return
	}

	err := h.srv.Update(mName, mType, value)
	if err != nil {
		http.Error(w, "тип или значение некорректно", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
