package config

import (
	"encoding/json"
	"errors"
	"os"
)

type BotConfig struct {
	TelegramTokenBot string `json:"telegram_bot_token"`
}

var cfg *BotConfig

func ReadConfig(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg = &BotConfig{}
	err = decoder.Decode(cfg)
	if err != nil {
		return
	}
}

func GetToken() (string, error){
	if cfg == nil{
		return "", errors.New("empty token")
	}
	return cfg.TelegramTokenBot, nil
}
