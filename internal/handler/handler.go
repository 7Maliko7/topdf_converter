package handler

import (
	"log"
	"topdf_converter/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "Отправьте мне изображения, чтобы создать PDF.")
	bot.Send(reply)
}

func HandleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет! Я твой PDF-бот помощник. Отправь /save, чтобы собрать PDF-файл")
	bot.Send(reply)
}

func HandleSavePDF(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	imgList, err := service.SaveMediaGroup(bot, msg)
	if err != nil {
		log.Fatal(err)
	}
	pdfpath := "newpdf"
	err = service.MakePDF(imgList, pdfpath)
	if err != nil {
		log.Fatal(err)
	}
	err = service.SendPDF(bot, msg.Chat.ID, pdfpath)
	if err != nil {
		log.Fatal(err)
	}
}
