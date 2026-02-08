package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type BotConfig struct {
	TelegramTokenBot   string `json:"telegram_bot_token"`
	TelegramBotDebug   bool   `json:"telegram_bot_debug"`
	TelegramBotTimeout int    `json:"telegram_bot_timeout_seconds"`
	PdfPath            string `json:"pdf_path"`
}

func ReadConfig(filename string) (*BotConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Println(closeErr)
		}
	}()
	decoder := json.NewDecoder(file)
	cfg := &BotConfig{}
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}
	err = cfg.validate()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (bc *BotConfig) validate() error {
	if bc.TelegramTokenBot == "" {
		return errors.New("no bot config")
	}
	if bc.TelegramBotTimeout == 0 {
		return errors.New("no bot timeout")
	}
	if bc.PdfPath == "" {
		return errors.New("no path for saving pdf")
	}
	return nil
}
