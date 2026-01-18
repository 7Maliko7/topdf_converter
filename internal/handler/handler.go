package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"topdf_converter/internal/service"
	"topdf_converter/pkg/metrics"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	photoBuffer *PhotoBuffer
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	photoBuffer := NewPhotoBuffer()
	return &Handler{
		bot:         bot,
		photoBuffer: photoBuffer,
	}
}

func (h *Handler) Start(chatID int64) {
	helloMsg := tgbotapi.NewMessage(chatID, "Привет, я твой бот-помощник по конвертации фото в pdf. Нажми кнопку \"Начать\", чтобы начать отправку фото")
	h.bot.Send(helloMsg)
}

func (h *Handler) Begin(chatID, userID int64) {
	buf := h.photoBuffer.Get(userID)
	if len(buf) > 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ты еще не закончил с предыдущим файлом. Нажми \"Завершить\" или \"Очистить\""))
		return
	}
	h.bot.Send(tgbotapi.NewMessage(chatID, "Отправь мне изображения, чтобы создать PDF. После того, как отправишь все нужные фото, нажми \"Завершить\""))
}

func (h *Handler) Clear(chatID, userID int64) {
	h.photoBuffer.Clear(userID)
	h.bot.Send(tgbotapi.NewMessage(chatID, "Фото удалены, можно начать создание нового файла"))
}

func (h *Handler) Done(chatID, userID int64) {
	defer h.photoBuffer.Clear(userID)

	h.bot.Send(tgbotapi.NewMessage(chatID, "Создаю файл в формате PDF. Это может занять немного времени"))

	fileIDs := h.photoBuffer.Get(userID)
	if len(fileIDs) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Нет фото в буфере"))
		return
	}

	dir := fmt.Sprintf("temp_photos_%d", userID)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Printf("Ошибка создания директории: %v", err)
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка сервера"))
		return
	}
	defer os.RemoveAll(dir)

	var paths []string
	for i, fileID := range fileIDs {
		localPath := filepath.Join(dir, fmt.Sprintf("img_%03d.jpg", i+1))
		file, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			log.Printf("GetFile error: %v", err)
			continue
		}
		url := file.Link(h.bot.Token)
		if err := service.DownloadPhoto(url, localPath); err != nil {
			log.Printf("Ошибка скачивания фото: %v", err)
			continue
		}
		paths = append(paths, localPath)
	}

	pdfPath := filepath.Join(dir, "result.pdf")
	if err := service.CreatePDFWithGoPDF(paths, pdfPath); err != nil {
		log.Printf("Ошибка создания PDF: %v", err)
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при генерации PDF."))
		os.RemoveAll(dir)
		return
	}

	f, err := os.Open(pdfPath)
	if err != nil {
		log.Printf("Ошибка открытия PDF: %v", err)
		h.bot.Send(tgbotapi.NewMessage(chatID, "Не удалось открыть PDF."))
		os.RemoveAll(dir)
		return
	}
	defer f.Close()

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   "photos.pdf",
		Reader: f,
	})
	for i := 0; i < 3; i++ {
		_, err := h.bot.Send(doc)
		if err == nil {
			metrics.PdfCount.Inc()
			h.bot.Send(tgbotapi.NewMessage(chatID, "Твой PDF готов."))
			return
		}
		tgErr, ok := err.(*tgbotapi.Error)
		if !ok {
			log.Printf("Ошибка отправки PDF: %v", err)
			h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при отправке PDF."))
			return
		}
		if tgErr.Code == 400 && tgErr.Message == "Bad Request: message text is empty" {
			continue
		}
		log.Printf("Ошибка отправки PDF: %v", err)
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при отправке PDF."))
		return
	}
}

func (h *Handler) NewFile(chatID, userID int64) {
	buf := h.photoBuffer.Get(userID)
	if len(buf) > 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ты еще не закончил с предыдущим файлом. Нажми \"Завершить\" или \"Очистить\""))
		return
	}
	h.bot.Send(tgbotapi.NewMessage(chatID, "Отправь мне изображения для нового файла. После того, как отправишь все нужные фото, нажми \"Завершить\""))
}

func (h *Handler) Help(chatID int64) {
	h.bot.Send(tgbotapi.NewMessage(chatID, "Если ты получил ошибку, нажми \"Новый файл\". \nЕсли хочешь перезапустить бот, отправь \"\\start\". \nЕсли хочешь удалить фото без создания файла, нажми \"Очистить\""))
}

func (h *Handler) Photo(update tgbotapi.Update, userID int64) {
	metrics.PhotoCount.Inc()
	largestPhoto := update.Message.Photo[len(update.Message.Photo)-1]
	h.photoBuffer.Add(userID, largestPhoto.FileID)
}
