package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"topdf_converter/internal/config"
	hndl "topdf_converter/internal/handler"
	"topdf_converter/internal/storage/tg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	path    string
}

func NewBot(cfg config.BotConfig) (*Bot, error) {
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
	kb.CreateButtons()

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
		return errors.New("No messages")
	}
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	// startTime := time.Now()
	// defer func() {
	// 	duration := time.Since(startTime).Seconds()
	// 	metrics.UserInteractionDuration.Observe(duration)
	// 	log.Printf("Interaction duration for user %d: %f seconds", chatID, duration)
	// }()

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	if len(update.Message.Photo) > 0 {
		handler.Photo(update, userID)
	} else {
		switch update.Message.Text {
		case "Начать":
			err := handler.Begin(chatID, userID)
			if err != nil {
				return err
			}
		case "Очистить":
			err := handler.Clear(chatID, userID)
			if err != nil {
				return err
			}
		case "Завершить":
			err := handler.Done(userID, chatID)
			if err != nil {
				return err
			}
		case "Новый файл":
			err := handler.NewFile(chatID, userID)
			if err != nil {
				return err
			}
		case "Помощь":
			err := handler.Help(chatID)
			if err != nil {
				return err
			}
		case "/start":
			err := handler.Start(chatID)
			if err != nil {
				return err
			}
		case "/help":
			err := handler.Help(chatID)
			if err != nil {
				return err
			}
		case "/done":
			err := handler.Done(userID, chatID)
			if err != nil {
				return err
			}
		case "/reset":
			err := handler.Clear(chatID, userID)
			if err != nil {
				return err
			}
		}
		b.bot.Send(msg)
	}
	return nil
}
