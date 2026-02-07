package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/7Maliko7/topdf_converter/internal/service"
	"github.com/7Maliko7/topdf_converter/internal/storage/tg"
	"github.com/7Maliko7/topdf_converter/pkg/metrics"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StartMsg       = "Привет, я твой бот-помощник по конвертации фото в pdf. Нажми кнопку \"Начать\", чтобы начать отправку фото"
	BufferMsg      = "Ты еще не закончил с предыдущим файлом. Нажми \"Завершить\" или \"Очистить\""
	BeginMsg       = "Отправь мне изображения, чтобы создать PDF. После того, как отправишь все нужные фото, нажми \"Завершить\""
	ClearMsg       = "Фото удалены, можно начать создание нового файла"
	ProcessPDFMsg  = "Создаю файл в формате PDF. Это может занять немного времени"
	EmptyBufferMsg = "Нет фото в буфере"
	PDFGenErrorMsg = "Ошибка при генерации PDF."
	NewFileMsg     = "Отправь мне изображения для нового файла. После того, как отправишь все нужные фото, нажми \"Завершить\""
	HelpMsg        = "Если ты получил ошибку, нажми \"Новый файл\". \nЕсли хочешь перезапустить бот, отправь \"\\start\". \nЕсли хочешь удалить фото без создания файла, нажми \"Очистить\""
	ServerErrorMsg = "Ошибка сервера"
)

type Handler struct {
	photoBuffer *photoBuffer
	storage     *tg.Storage
	path        string
}

func NewHandler(path string, st *tg.Storage) *Handler {
	photoBuffer := newPhotoBuffer()
	return &Handler{
		photoBuffer: photoBuffer,
		path:        path,
		storage:     st,
	}
}

func (h *Handler) Start() string {
	return StartMsg
}

func (h *Handler) Begin(userID int64) string {
	buf := h.photoBuffer.get(userID)
	if len(buf) > 0 {
		return BufferMsg
	}
	return BeginMsg
}

func (h *Handler) Clear(userID int64) string {
	h.photoBuffer.clear(userID)
	return ClearMsg
}

func (h *Handler) Done(userID, chatID int64) (*tgbotapi.DocumentConfig, string, error) {
	defer h.photoBuffer.clear(userID)

	fileIDs := h.photoBuffer.get(userID)
	if len(fileIDs) == 0 {
		return nil, EmptyBufferMsg, nil
	}

	dir := fmt.Sprintf("%s_%d", h.path, userID)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Printf("Ошибка создания директории: %v", err)
		return nil, ServerErrorMsg, err
	}
	defer os.RemoveAll(dir)

	var paths []string
	for i, fileID := range fileIDs {
		localPath := filepath.Join(dir, fmt.Sprintf("img_%03d.jpg", i+1))
		file, err := h.storage.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			log.Printf("GetFile error: %v", err)
			continue
		}
		if err := service.DownloadPhoto(file.Url, localPath); err != nil {
			log.Printf("Ошибка скачивания фото: %v", err)
			continue
		}
		paths = append(paths, localPath)
	}

	pdfPath := filepath.Join(dir, "result.pdf")
	if err := service.CreatePDFWithGoPDF(paths, pdfPath); err != nil {
		log.Printf("Ошибка создания PDF: %v", err)
		os.RemoveAll(dir)
		return nil, PDFGenErrorMsg, err
	}

	doc, docMsg, err := h.storage.CreateDocument(chatID, pdfPath)
	if err != nil {
		os.RemoveAll(dir)
		return nil, docMsg, err
	}

	return doc, docMsg, nil
}

func (h *Handler) Process() string {
	return ProcessPDFMsg
}

func (h *Handler) NewFile(userID int64) string {
	buf := h.photoBuffer.get(userID)
	if len(buf) > 0 {
		return BufferMsg
	}
	return NewFileMsg
}

func (h *Handler) Help() string {
	return HelpMsg
}

func (h *Handler) Photo(update tgbotapi.Update, userID int64) {
	metrics.PhotoCount.Inc()
	sourcePhoto := update.Message.Photo[len(update.Message.Photo)-1]
	h.photoBuffer.add(userID, sourcePhoto.FileID)
}
