package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/AA122AA/metring/internal/agent"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		fmt.Println("got interruption, cancelling ctx")
		cancel()
	}()

	cfg, err := agent.Read("")
	if err != nil {
		fmt.Printf("can not read config - %v", err)
		return
	}

	var wg sync.WaitGroup

	mAgent := agent.NewMetricAgent(cfg)
	go mAgent.Run(ctx, &wg)
	wg.Add(1)
	fmt.Println("Ran agent")

	client := agent.NewMetricClient(mAgent, cfg)
	go client.Run(ctx, &wg)
	wg.Add(1)
	fmt.Println("Ran client")

	wg.Wait()
}
