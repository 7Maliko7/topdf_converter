package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	bot "topdf_converter/internal/bot"
	"topdf_converter/internal/config"
	"topdf_converter/pkg/metrics"
)

//TODO
//Контролируемая остановка сервера с метриками (грейсфул шатдаун)

func main() {
	config.ReadConfig("tokencfg.json")
	token, err := config.GetToken()
	if err != nil {
		return
	}
	//goroutine
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	errChan := make(chan error)

	bot := bot.NewBot(token)

	bot.Start(ctx, token)

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
