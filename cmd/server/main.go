// пакеты исполняемых приложений должны называться main
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/AA122AA/metring/internal/server"
	"github.com/AA122AA/metring/internal/server/config"
	"github.com/AA122AA/metring/internal/server/database"
	"github.com/AA122AA/metring/internal/server/database/query"
	mHandler "github.com/AA122AA/metring/internal/server/handler"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service/metrics"
	"github.com/AA122AA/metring/internal/server/service/saver"
	"github.com/AA122AA/metring/internal/zapcfg"
	"github.com/creasty/defaults"
	"github.com/go-faster/sdk/zctx"
	_ "github.com/jackc/pgx"
	"go.uber.org/zap"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	if err := run(); err != nil {
		if err.Error() == "http: Server closed" {
			return
		}
		log.Fatalf("got error in run func: %v", err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	// Main logger
	lg, err := zapcfg.New().Build()
	if err != nil {
		log.Fatalf("got error while creating logger - %v", err)
	}

	// Logger flusher
	flush := func() { _ = lg.Sync() }
	defer flush()

	// Panic recover
	defer func() {
		if r := recover(); r != nil {
			lg.Fatal("Panic recovering", zap.Any("panic", r))
			os.Exit(2)
		}
	}()

	// Main context.
	ctx, cancel := signal.NotifyContext(zctx.Base(context.Background(), lg), os.Interrupt)
	defer cancel()

	// Reading config
	cfg := &config.Config{}
	if err := defaults.Set(cfg); err != nil {
		lg.Fatal("error setting defaults for config", zap.Error(err))
	}
	cfg.ParseConfig()
	cfg.LoadEnv()

	lg.Debug(
		"server config",
		zap.String("address", cfg.HostAddr),
		zap.String("template path", cfg.TemplatePath),
		zap.String("file storage path", cfg.SaverCfg.FileStoragePath),
		zap.Int("store interval", cfg.SaverCfg.StoreInterval),
		zap.Bool("restore", cfg.SaverCfg.Restore),
	)

	// Init repo
	var repo repository.MetricsRepository
	repo = repository.NewMemStorage()

	var dBase *database.Database
	// Init DB
	if cfg.DatabaseDSN != "" {
		dBase = database.New(ctx, "pgx", cfg.DatabaseDSN)
		err = dBase.Ping(ctx)
		if err != nil {
			lg.Fatal("can not connect to db", zap.Error(err))
		}
		defer dBase.Close()

		// Run migrations
		err = dBase.Migrate(ctx)
		if err != nil {
			lg.Fatal("can not run migrations", zap.Error(err))
		}
		d := dBase.DB()
		q := query.New(d)
		repo = repository.NewPSQLStorage(ctx, q, dBase)
	}

	// Create wait group
	var wg sync.WaitGroup

	// Init services
	srv := metrics.NewMetrics(ctx, repo)

	var saverSvc *saver.Saver
	if cfg.SaverCfg.FileStoragePath != "" {
		lg.Debug("file storage path is not empty", zap.String("FileStoragePath", cfg.SaverCfg.FileStoragePath))
		saverSvc = saver.NewSaver(ctx, cfg.SaverCfg, repo)
		wg.Add(1)
		go saverSvc.Run(ctx, &wg)
		lg.Debug("Ran saver")
	}
	if saverSvc == nil {
		lg.Debug("saver should be nil", zap.Any("saverSvc", saverSvc))
	}

	// Init handlers
	metricHandler := mHandler.NewMetricsHandler(ctx, cfg.TemplatePath, srv, saverSvc)
	pingHandler := mHandler.NewPingHandler(ctx, dBase)

	// Init routers
	router := server.NewRouter(ctx, metricHandler, pingHandler)

	// Init server
	server := server.NewServer(ctx, cfg, router)

	wg.Add(1)
	go server.OnShutDown(ctx, &wg)
	lg.Debug("Ran On ShutDown")

	err = server.Run(ctx)

	wg.Wait()

	return err
}
