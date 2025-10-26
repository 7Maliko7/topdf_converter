package main

import (
	"topdf_converter/internal/handler"
	"topdf_converter/internal/config"
)

func main() {
	config.ReadConfig("tokencfg.json")
	token, err := config.GetToken()
	if err != nil {
		return
	}

	handler.Start(token)

}
