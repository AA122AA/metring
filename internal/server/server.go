package server

import (
	"fmt"
	"net/http"

	"github.com/AA122AA/metring/internal/server/config"
	"github.com/AA122AA/metring/internal/server/handler"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	srv *http.Server
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		srv: &http.Server{
			Addr:    cfg.HostAddr,
			Handler: router(),
		},
	}
}

func (s *Server) Run() error {
	fmt.Printf("Start server on %v\n", s.srv.Addr)
	return s.srv.ListenAndServe()
}

func router() *chi.Mux {
	repo := repository.NewMemStorage()
	srv := service.NewMetrics(repo)
	h := handler.NewMetricsHandler(srv)

	router := chi.NewRouter()
	router.Get("/", h.All)
	router.Get("/value/{mType}/{mName}", h.Get)
	router.Post("/update/{mType}/{mName}/{value}", h.Update)

	return router
}
