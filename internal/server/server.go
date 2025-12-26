package server

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/AA122AA/metring/internal/server/config"
	"github.com/AA122AA/metring/internal/server/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type metricsHandler interface {
	All(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	GetJSON(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	UpdateJSON(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	srv *http.Server
	lg  *zap.Logger
}

func NewServer(ctx context.Context, cfg *config.Config, router http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:    cfg.HostAddr,
			Handler: router,
		},
		lg: zctx.From(ctx).Named("server"),
	}
}

func (s *Server) OnShutDown(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	if s.srv != nil {
		if err := s.srv.Shutdown(ctx); err != nil {
			s.lg.Fatal("failed to shutdown http server", zap.Error(err))
		}
		s.lg.Info("shutdown http server")
	}
}

func (s *Server) Run(ctx context.Context) error {
	port := ":" + strings.Split(s.srv.Addr, ":")[1]
	listener, err := net.Listen("tcp", port)
	if err != nil {
		s.lg.Fatal("failed to create listener", zap.Error(err))
	}
	s.lg.Debug(listener.Addr().String())

	s.lg.Info("Start server on", zap.String("addr", s.srv.Addr))

	return s.srv.Serve(listener)
}

func NewRouter(ctx context.Context, h metricsHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", middleware.Wrap(
		middleware.Wrap(
			http.HandlerFunc(h.All),
			middleware.WithLogger(zctx.From(ctx).Named("GetAll"))),
		middleware.WithCompression()),
	)
	router.Route("/value/", func(r chi.Router) {
		r.Post("/", middleware.Wrap(
			middleware.Wrap(
				http.HandlerFunc(h.GetJSON),
				middleware.WithLogger(zctx.From(ctx).Named("GetValueJSON"))),
			middleware.WithCompression()),
		)
		r.Get("/{mType}/{mName}", middleware.Wrap(
			middleware.Wrap(
				http.HandlerFunc(h.Get),
				middleware.WithLogger(zctx.From(ctx).Named("GetValue"))),
			middleware.WithCompression()),
		)
	})
	router.Route("/update", func(r chi.Router) {
		r.Post("/", middleware.Wrap(
			middleware.Wrap(
				http.HandlerFunc(h.UpdateJSON),
				middleware.WithLogger(zctx.From(ctx).Named("UpdateValueJSON"))),
			middleware.WithCompression()),
		)
		r.Post("/{mType}/{mName}/{value}", middleware.Wrap(
			middleware.Wrap(
				http.HandlerFunc(h.Update),
				middleware.WithLogger(zctx.From(ctx).Named("UpdateValue"))),
			middleware.WithCompression()),
		)
	})

	return router
}
