package main

import (
	"pdf-bot/internal/handler"
	"pdf-bot/internal/config"
)

func main() {
	config.ReadConfig("tokencfg.json")
	token, err := config.GetToken()
	if err != nil {
		return
	}

	handler.Start(token)

}
