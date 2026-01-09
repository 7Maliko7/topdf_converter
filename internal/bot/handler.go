package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct{
	bot *tgbotapi.BotAPI
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler{
	return &Handler{
		bot: bot,
	}
}

func(h *Handler) Start(chatID int64, kb *Keyboard) {
	helloMsg := tgbotapi.NewMessage(chatID, "Привет, я твой бот-помощник по конвертации фото в pdf. Нажми кнопку \"Начать\", чтобы начать отправку фото")
	helloMsg.ReplyMarkup = kb.GetMarkup()
	h.bot.Send(helloMsg)
}

func(h *Handler) Begin(chatID int64){}

func(h *Handler) Clear(chatID int64){}

func(h *Handler) Done(chatID int64){}

func(h *Handler) NewFile(chatID int64){}

func(h *Handler) Help(chatID int64) {
	h.bot.Send(tgbotapi.NewMessage(chatID, "Если ты получил ошибку, нажми \"Новый файл\". \nЕсли хочешь перезапустить бот, отправь \"\\start\". \nЕсли хочешь удалить фото без создания файла, нажми \"Очистить\""))
}


