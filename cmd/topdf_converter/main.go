package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	wg := &sync.WaitGroup{}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bot.Start(ctx, token); err != nil {
			log.Printf("bot stopped with error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := metrics.StartMetricsServer(ctx); err != nil {
			log.Printf("prometheus stopped with error: %v", err)
		}
	}()

	<-sig
	log.Println("shutdown signal received")

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("shutdown completed")
	case <-time.After(10 * time.Second):
		log.Println("shutdown timeout")
	}

}
