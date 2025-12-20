package server

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/AA122AA/metring/internal/server/config"
	"github.com/AA122AA/metring/internal/server/handler"
	"github.com/AA122AA/metring/internal/server/middleware"
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
	port := ":" + strings.Split(s.srv.Addr, ":")[1]
	listener, err := net.Listen("tcp", port)
	if err != nil {
		s.lg.Fatal("failed to create listener", zap.Error(err))
	}
	s.lg.Info(listener.Addr().String())

	s.lg.Info("Start server on", zap.String("addr", s.srv.Addr))

	return s.srv.Serve(listener)
}

func router(ctx context.Context, tPath string) *chi.Mux {
	repo := repository.NewMemStorage()
	srv := service.NewMetrics(ctx, repo)
	h := handler.NewMetricsHandler(ctx, tPath, srv)

	router := chi.NewRouter()
	router.Get("/", middleware.Wrap(
		http.HandlerFunc(h.All),
		middleware.WithLogger(zctx.From(ctx).Named("GetAll"))),
	)
	router.Route("/value/", func(r chi.Router) {
		r.Post("/", middleware.Wrap(
			http.HandlerFunc(h.GetJSON),
			middleware.WithLogger(zctx.From(ctx).Named("GetValueJSON"))),
		)
		r.Get("/{mType}/{mName}", middleware.Wrap(
			http.HandlerFunc(h.Get),
			middleware.WithLogger(zctx.From(ctx).Named("GetValue"))),
		)
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", middleware.Wrap(
			http.HandlerFunc(h.UpdateJSON),
			middleware.WithLogger(zctx.From(ctx).Named("UpdateValueJSON"))),
		)
		r.Post("/{mType}/{mName}/{value}", middleware.Wrap(
			http.HandlerFunc(h.Update),
			middleware.WithLogger(zctx.From(ctx).Named("UpdateValue"))),
		)
	})

	return router
}
