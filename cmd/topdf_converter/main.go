package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	bot "github.com/7Maliko7/topdf_converter/internal/bot"
	"github.com/7Maliko7/topdf_converter/internal/config"
	"github.com/7Maliko7/topdf_converter/pkg/metrics"
)

func main() {
	botCfg := flag.String("config", "./configs/bot_config.json", "config path")
	flag.Parse()

	if *botCfg == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.ReadConfig(*botCfg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	errChan := make(chan error)

	bot, err := bot.NewBot(cfg)
	if err != nil {
		log.Println(err)
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
