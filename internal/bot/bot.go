package bot

import (
	"context"
	"fmt"
	"log"
	hndl "topdf_converter/internal/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//TODO
/*
Не начинать новый файл до очистки буфера
Помощь прописать
Растягивание картинок убрать
Решить вопрос с не фотографиями
*/

//var userState = make(map[int64]string)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(token string) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	fmt.Printf("Авторизация прошла успешно. Бот работает в режиме %s\n", bot.Self.UserName)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return &Bot{
		bot: bot,
	}
}

func (b *Bot) Start(ctx context.Context, token string) {
	handler := hndl.NewHandler(b.bot)

	kb := NewKeyboard()
	kb.CreateButtons()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("stop updates")
				b.bot.StopReceivingUpdates()
				return
			case update := <-updates:
				CheckUpdates(b.bot, update, handler, kb)
			}
		}
	}()

}

func CheckUpdates(bot *tgbotapi.BotAPI, update tgbotapi.Update, handler *hndl.Handler, kb *Keyboard) {

	if update.Message == nil {
		return
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
			handler.Begin(chatID, userID)
		case "Очистить":
			handler.Clear(chatID, userID)
		case "Завершить":
			handler.Done(chatID, userID)
		case "Новый файл":
			handler.NewFile(chatID,userID)
		case "Помощь":
			handler.Help(chatID)
		case "/start":
			handler.Start(chatID)
		case "/help":
			handler.Help(chatID)
		case "/done":
			handler.Done(chatID, userID)
		case "/reset":
			handler.Clear(chatID, userID)
		}
		bot.Send(msg)
	}
}
