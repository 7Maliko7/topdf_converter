package bot

import (
	"context"
	"fmt"
	"log"

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

func Start(ctx context.Context, token string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	fmt.Printf("Авторизация прошла успешно. Бот работает в режиме %s\n", bot.Self.UserName)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	photoBuffer := NewPhotoBuffer()
	handler := NewHandler(bot, photoBuffer)

	kb := NewKeyboard()
	kb.CreateButtons()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	go func() {
		<-ctx.Done()
		bot.StopReceivingUpdates()
	}()

	for update := range updates {
		select {
		case <-ctx.Done():
			return nil
		default:
			CheckUpdates(bot, update, handler, kb)
		}

	}
	return nil
}

func CheckUpdates(bot *tgbotapi.BotAPI, update tgbotapi.Update, handler *Handler, kb *Keyboard) {

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
			handler.Begin(chatID)
		case "Очистить":
			handler.Clear(chatID, userID)
		case "Завершить":
			handler.Done(chatID, userID)
		case "Новый файл":
			handler.NewFile(chatID)
		case "Помощь":
			handler.Help(chatID)
		case "/start":
			handler.Start(chatID, kb)
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
