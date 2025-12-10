package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/AA122AA/metring/internal/agent"
	"github.com/AA122AA/metring/internal/zapcfg"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

func main() {
	lg, err := zapcfg.New().Build()
	if err != nil {
		panic(err)
	}

	// Logger flusher
	flush := func() { _ = lg.Sync() }
	defer flush()

	ctx, cancel := signal.NotifyContext(zctx.Base(context.Background(), lg), os.Interrupt)
	defer func() {
		lg.Info("got interruption, cancelling ctx")
		cancel()
	}()

	cfg, err := agent.Read("")
	if err != nil {
		lg.Error("got interruption, cancelling ctx", zap.Error(err))
		return
	}

	cfg.ParseFlags(ctx)

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
