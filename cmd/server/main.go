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
	mHandler "github.com/AA122AA/metring/internal/server/handler"
	"github.com/AA122AA/metring/internal/server/repository"
	"github.com/AA122AA/metring/internal/server/service/metrics"
	"github.com/AA122AA/metring/internal/server/service/saver"
	"github.com/AA122AA/metring/internal/zapcfg"
	"github.com/creasty/defaults"
	"github.com/go-faster/sdk/zctx"
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

	// Init services
	repo := repository.NewMemStorage()
	srv := metrics.NewMetrics(ctx, repo)
	saver := saver.NewSaver(ctx, cfg.SaverCfg, repo)

	// Init handlers
	metricHandler := mHandler.NewMetricsHandler(ctx, cfg.TemplatePath, srv, saver)

	// Init routers
	router := server.NewRouter(ctx, metricHandler)

	// Init server
	server := server.NewServer(ctx, cfg, router)

	var wg sync.WaitGroup
	wg.Add(1)
	go saver.Run(ctx, &wg)
	lg.Debug("Ran saver")

	wg.Add(1)
	go server.OnShutDown(ctx, &wg)
	lg.Debug("Ran On ShutDown")

	err = server.Run(ctx)

	wg.Wait()

	return err
}
