package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Keyboard struct {
	Buttons [][]tgbotapi.KeyboardButton
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		Buttons: [][]tgbotapi.KeyboardButton{},
	}
}

func (k *Keyboard) AddRow(buttons ...tgbotapi.KeyboardButton) {
	k.Buttons = append(k.Buttons, buttons)
}

func (k *Keyboard) GetMarkup() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       k.Buttons,
		ResizeKeyboard: true,
	}
}

func (k *Keyboard) CreateButtonTemplate() {
	k.AddRow(tgbotapi.NewKeyboardButton("Начать"))
	k.AddRow(tgbotapi.NewKeyboardButton("Очистить"))
	k.AddRow(tgbotapi.NewKeyboardButton("Завершить"))
	k.AddRow(tgbotapi.NewKeyboardButton("Новый файл"))
	k.AddRow(tgbotapi.NewKeyboardButton("Помощь"))
}
