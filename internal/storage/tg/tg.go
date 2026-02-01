package tg

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Storage struct {
	Bot *tgbotapi.BotAPI
}

type TgFile struct {
	File *tgbotapi.File
	Url  string
}

func NewStorage(bot *tgbotapi.BotAPI) *Storage {
	return &Storage{
		Bot: bot,
	}
}

func (s *Storage) GetFile(config tgbotapi.FileConfig) (*TgFile, error) {
	file, err := s.Bot.GetFile(config)
	if err != nil {
		return nil, err
	}
	return &TgFile{
		File: &file,
		Url:  file.Link(s.Bot.Token),
	}, nil
}

func (s *Storage) SendDocument(doc tgbotapi.DocumentConfig) error {
	_, err := s.Bot.Send(doc)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) SendMessage(chatID int64, msg string) error {
	_, err := s.Bot.Send(tgbotapi.NewMessage(chatID, msg))
	if err != nil {
		return err
	}
	return nil
}
