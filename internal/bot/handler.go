package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"topdf_converter/pkg/metrics"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	photoBuffer *PhotoBuffer
}

func NewHandler(bot *tgbotapi.BotAPI, photoBuffer *PhotoBuffer) *Handler {
	return &Handler{
		bot:         bot,
		photoBuffer: photoBuffer,
	}
}

func (h *Handler) Start(chatID int64, kb *Keyboard) {
	helloMsg := tgbotapi.NewMessage(chatID, "Привет, я твой бот-помощник по конвертации фото в pdf. Нажми кнопку \"Начать\", чтобы начать отправку фото")
	helloMsg.ReplyMarkup = kb.GetMarkup()
	h.bot.Send(helloMsg)
}

func (h *Handler) Begin(chatID int64) {
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
		if err := downloadPhoto(url, localPath); err != nil {
			log.Printf("Ошибка скачивания фото: %v", err)
			continue
		}
		paths = append(paths, localPath)
	}

	pdfPath := filepath.Join(dir, "result.pdf")
	if err := createPDFWithGoPDF(paths, pdfPath); err != nil {
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
	if _, err := h.bot.Send(doc); err != nil {
		// if errors.As(err, "Bad Request: message text is empty") {

		// }
		//TODO is err = timeout "error_code":400,"description":"Bad Request: message text is empty" повторить оправку, сообщение об отпрвке
		log.Printf("Ошибка отправки PDF: %v", err)
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при отправке PDF."))
	}
	metrics.PdfCount.Inc()
}

func (h *Handler) NewFile(chatID int64) {
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
