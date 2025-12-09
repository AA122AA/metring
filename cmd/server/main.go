// пакеты исполняемых приложений должны называться main
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/AA122AA/metring/internal/server"
	"github.com/AA122AA/metring/internal/server/config"
	"github.com/AA122AA/metring/internal/zapcfg"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	if err := run(); err != nil {
		log.Fatalf("got error in run func - %v", err)
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

	ctx, cancel := signal.NotifyContext(zctx.Base(context.Background(), lg), os.Interrupt)
	defer cancel()

	cfg := &config.Config{}
	cfg.ParseConfig(ctx)

	server := server.NewServer(ctx, cfg)

	return server.Run()
}
