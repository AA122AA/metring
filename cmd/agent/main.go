package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/AA122AA/metring/internal/agent"
	"github.com/AA122AA/metring/internal/zapcfg"
	"github.com/caarlos0/env"
	"github.com/creasty/defaults"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

func main() {
	lg, err := zapcfg.New().Build()
	if err != nil {
		log.Fatalf("err while building zapcfg: %v", err)
	}

	// Logger flusher
	flush := func() { _ = lg.Sync() }
	defer flush()

	ctx, cancel := signal.NotifyContext(zctx.Base(context.Background(), lg), os.Interrupt)
	defer func() {
		lg.Info("got interruption, cancelling ctx")
		cancel()
	}()

	cfg := &agent.Config{}
	if err := defaults.Set(cfg); err != nil {
		lg.Fatal("error setting defaults for config", zap.Error(err))
	}

	// cfg, err := agent.Read("")
	// if err != nil {
	// 	lg.Fatal("got interruption, cancelling ctx", zap.Error(err))
	// }

	cfg.ParseFlags()

	if err = env.Parse(cfg); err != nil {
		lg.Fatal("error setting config from env", zap.Error(err))
	}

	lg.Debug(
		"config values",
		zap.String("address", cfg.URL),
		zap.Int("report interval", cfg.ReportInterval),
		zap.Int("poll interval", cfg.PollInterval),
	)

	var wg sync.WaitGroup

	mAgent := agent.NewMetricAgent(ctx, cfg)
	go mAgent.Run(ctx, &wg)
	wg.Add(1)
	lg.Info("Ran agent")

	client := agent.NewMetricClient(ctx, mAgent, cfg)
	go client.Run(ctx, &wg)
	wg.Add(1)
	lg.Info("Ran client")

	wg.Wait()
}
