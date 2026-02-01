package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	bot "topdf_converter/internal/bot"
	"topdf_converter/internal/config"
	"topdf_converter/pkg/metrics"
)

func main() {
	cfg, err := config.ReadConfig("tokencfg.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	errChan := make(chan error)

	bot, err := bot.NewBot(*cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	bot.Start(ctx)

	metrics.StartMetricsServer(ctx, errChan)
	select {
	case s := <-sig:
		cancel()
		log.Printf("shutdown completed: %v\n", s)
	case er := <-errChan:
		cancel()
		log.Fatal(er)
	}
}
