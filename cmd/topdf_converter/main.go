package main

import (
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
	metrics.StartMetricsServer()
	bot.Start(token)

}
