package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/signintech/gopdf"
)

var userState = make(map[int64]string)

func Start(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	fmt.Printf("Авторизация прошла успешно. Бот работает в режиме %s\n", bot.Self.UserName)

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	photoBuffer := NewPhotoBuffer()

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		// if !update.Message.IsCommand(){
		// 	continue
		// }
		state, ok := userState[userID]
		if ok && state == "waiting_for_photo" {
			HandleSavePDF(bot, update.Message)
			msg.Text = "take your pdf"
			bot.Send(msg)
			return
		}

		switch {
		case update.Message.IsCommand():
			switch update.Message.Command() {
			case "start":
				HandleStart(bot, update.Message)
			case "help":
				HandleHelp(bot, update.Message)
			case "save":
				userState[userID] = "waiting_for_photo"
				msg.Text = "do send photos"
			case "done":
				fileIDs := photoBuffer.Get(userID)
				if len(fileIDs) == 0 {
					bot.Send(tgbotapi.NewMessage(chatID, "Нет фото в буфере."))
					continue
				}

				// Создаём временную папку
				dir := fmt.Sprintf("temp_photos_%d", userID)
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					log.Println("Ошибка создания директории:", err)
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка сервера."))
					continue
				}

				var paths []string
				for i, fileID := range fileIDs {
					localPath := filepath.Join(dir, fmt.Sprintf("img_%03d.jpg", i+1))
					if err := downloadPhoto(bot, fileID, localPath); err != nil {
						log.Println("Ошибка скачивания фото:", err)
						// можно продолжить дальше, пропуская этот файл
						continue
					}
					paths = append(paths, localPath)
				}

				pdfPath := filepath.Join(dir, "result.pdf")
				if err := createPDFWithGoPDF(paths, pdfPath); err != nil {
					log.Println("Ошибка создания PDF:", err)
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при генерации PDF."))
					// возможно очистка
					// os.RemoveAll(dir)
					continue
				}

				f, err := os.Open(pdfPath)
				if err != nil {
					log.Println("Ошибка открытия PDF:", err)
					bot.Send(tgbotapi.NewMessage(chatID, "Не удалось открыть PDF."))
					// очистка
					os.RemoveAll(dir)
					continue
				}
				defer f.Close()

				doc := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
					Name:   "photos.pdf",
					Reader: f,
				})
				if _, err := bot.Send(doc); err != nil {
					log.Println("Ошибка отправки PDF:", err)
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при отправке PDF."))
				}

				// Очистка после
				photoBuffer.Clear(userID)
				os.RemoveAll(dir)

			case "reset":
				photoBuffer.Clear(userID)
				msg.Text = "buffer cleared"
			default:
				msg.Text = "Unknown command"
			}
		case len(update.Message.Photo) > 0:
			largestPhoto := update.Message.Photo[len(update.Message.Photo)-1]
			photoBuffer.Add(userID, largestPhoto.FileID)
			msg.Text = "photo added"
			bot.Send(msg)
		default:
			bot.Send(tgbotapi.NewMessage(chatID, "add photo or done"))
		}

		bot.Send(msg)
	}
}

func downloadPhoto(bot *tgbotapi.BotAPI, fileID, savePath string) error {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return fmt.Errorf("GetFile error: %w", err)
	}
	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http.Get error: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("os.Create error: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("io.Copy error: %w", err)
	}
	return nil
}

func createPDFWithGoPDF(images []string, outputPath string) error {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	for _, img := range images {
		pdf.AddPage()
		// Вставляем изображение на весь лист
		if err := pdf.Image(img, 0, 0, gopdf.PageSizeA4); err != nil {
			return fmt.Errorf("ошибка вставки изображения %s: %w", img, err)
		}
	}

	if err := pdf.WritePdf(outputPath); err != nil {
		return fmt.Errorf("WritePdf error: %w", err)
	}
	return nil
}
