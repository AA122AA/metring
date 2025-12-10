package server

import (
	"context"
	"net/http"

	"github.com/AA122AA/metring/internal/server/config"
	"github.com/AA122AA/metring/internal/server/handler"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type Server struct {
	srv *http.Server
	lg  *zap.Logger
}

func NewServer(ctx context.Context, cfg *config.Config) *Server {
	return &Server{
		srv: &http.Server{
			Addr:    cfg.HostAddr,
			Handler: router(ctx, cfg.TemplatePath),
		},
		lg: zctx.From(ctx).Named("server"),
	}
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		if s.srv != nil {
			if err := s.srv.Shutdown(ctx); err != nil {
				s.lg.Fatal("failed to shutdown http server", zap.Error(err))
			}
			s.lg.Info("shutdown http server")
		}
	}()

	s.lg.Info("Start server on", zap.String("addr", s.srv.Addr))

	return s.srv.ListenAndServe()
}

func router(ctx context.Context, tPath string) *chi.Mux {
	repo := repository.NewMemStorage()
	srv := service.NewMetrics(ctx, repo)
	h := handler.NewMetricsHandler(ctx, tPath, srv)

	router := chi.NewRouter()
	router.Get("/", h.All)
	router.Get("/value/{mType}/{mName}", h.Get)
	router.Post("/update/{mType}/{mName}/{value}", h.Update)

	return router
}
