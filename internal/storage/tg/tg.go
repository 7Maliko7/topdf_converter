package tg

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const PDFOpenErrorMsg = "Не удалось открыть PDF."

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

func (s *Storage) CreateDocument(chatID int64, pdfPath string) (*tgbotapi.DocumentConfig, string, error) {
	f, err := os.Open(pdfPath)
	if err != nil {
		log.Printf("Ошибка открытия PDF: %v", err)
		return nil, PDFOpenErrorMsg, err
	}
	defer f.Close()

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   "photos.pdf",
		Reader: f,
	})
	return &doc, "", nil
}
