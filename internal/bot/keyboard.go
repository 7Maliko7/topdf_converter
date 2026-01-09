package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Keyboard struct {
	Buttons [][]tgbotapi.KeyboardButton
}

func NewKeyboard()*Keyboard{
	return &Keyboard{
		Buttons: [][]tgbotapi.KeyboardButton{},
	}
}

func(k *Keyboard) AddRow(buttons ...tgbotapi.KeyboardButton){
	k.Buttons = append(k.Buttons, buttons)
}

func CreateButton(text string)tgbotapi.KeyboardButton{
	return tgbotapi.NewKeyboardButton(text)
}

func(k *Keyboard) GetMarkup() tgbotapi.ReplyKeyboardMarkup{
	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard: k.Buttons,
		ResizeKeyboard: true,
	}
}
