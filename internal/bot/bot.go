package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/7Maliko7/topdf_converter/internal/config"
	hndl "github.com/7Maliko7/topdf_converter/internal/handler"
	"github.com/7Maliko7/topdf_converter/internal/storage/tg"
	"github.com/7Maliko7/topdf_converter/pkg/metrics"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	PDFSendErrorMsg = "Ошибка при отправке PDF."
	ReadyPDFMsg     = "Твой PDF готов."
)

type Bot struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	path    string
}

func NewBot(cfg *config.BotConfig) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramTokenBot)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Авторизация прошла успешно. Бот: %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)

	u.Timeout = cfg.TelegramBotTimeout
	updates := bot.GetUpdatesChan(u)

	bot.Debug = cfg.TelegramBotDebug

	return &Bot{
		bot:     bot,
		updates: updates,
		path:    cfg.PdfPath,
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	storage := tg.NewStorage(b.bot)
	handler := hndl.NewHandler(b.path, storage)

	kb := NewKeyboard()
	kb.CreateButtonTemplate()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("stop updates")
				b.bot.StopReceivingUpdates()
				return
			case update := <-b.updates:
				err := b.CheckUpdates(update, handler)
				log.Println(err)
			}
		}
	}()

}

func (b *Bot) CheckUpdates(update tgbotapi.Update, handler *hndl.Handler) error {
	if update.Message == nil {
		log.Println(errors.New("no messages"))
		return nil
	}
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	if len(update.Message.Photo) > 0 {
		handler.Photo(update, userID)
	} else {
		switch update.Message.Text {
		case "Начать":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Begin(userID)))
			return err
		case "Очистить":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Clear(userID)))
			return err
		case "Завершить":
			return b.Done(chatID, userID, handler)
		case "Новый файл":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.NewFile(userID)))
			return err
		case "Помощь":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Help()))
			return err
		case "/start":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Start()))
			return err
		case "/help":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Help()))
			return err
		case "/done":
			return b.Done(chatID, userID, handler)
		case "/reset":
			_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Clear(userID)))
			return err
		}
		_, err := b.bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) Done(chatID, userID int64, handler *hndl.Handler) error {
	_, err := b.bot.Send(tgbotapi.NewMessage(chatID, handler.Process()))
	if err != nil {
		return err
	}
	doc, docMsg, err := handler.Done(userID, chatID)
	if err != nil {
		_, err := b.bot.Send(tgbotapi.NewMessage(chatID, docMsg))
		if err != nil {
			return err
		}
		return err
	}

	err = b.SendDocument(*doc, chatID)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) SendDocument(doc tgbotapi.DocumentConfig, chatID int64) error {
	for {
		_, err := b.bot.Send(doc)
		if err != nil {
			tgErr, ok := err.(*tgbotapi.Error)
			if ok {
				log.Printf("Ошибка отправки PDF: %v", err)
				_, err = b.bot.Send(tgbotapi.NewMessage(chatID, PDFSendErrorMsg))
				if err != nil {
					log.Printf("Ошибка отправки сообщения: %v", err)
					return err
				}
				return nil
			}
			if tgErr != nil {
				log.Println(tgErr)
				if tgErr.Code == 400 || tgErr.Message == "Bad Request: message text is empty" {
					continue
				}
			}
		}
		metrics.PdfCount.Inc()
		_, err = b.bot.Send(tgbotapi.NewMessage(chatID, ReadyPDFMsg))
		if err != nil {
			return err
		}
		return nil
	}
}
