package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"topdf_converter/internal/service"
	"topdf_converter/internal/storage/tg"
	"topdf_converter/pkg/metrics"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func (h *Handler) Start(chatID int64) error {
	err := h.storage.SendMessage(chatID, "Привет, я твой бот-помощник по конвертации фото в pdf. Нажми кнопку \"Начать\", чтобы начать отправку фото")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Begin(chatID, userID int64) error {
	buf := h.photoBuffer.get(userID)
	if len(buf) > 0 {
		err := h.storage.SendMessage(chatID, "Ты еще не закончил с предыдущим файлом. Нажми \"Завершить\" или \"Очистить\"")
		if err != nil {
			return err
		}
		return nil
	}
	err := h.storage.SendMessage(chatID, "Отправь мне изображения, чтобы создать PDF. После того, как отправишь все нужные фото, нажми \"Завершить\"")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Clear(chatID, userID int64) error {
	h.photoBuffer.clear(userID)
	err := h.storage.SendMessage(chatID, "Фото удалены, можно начать создание нового файла")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Done(userID, chatID int64) error {
	defer h.photoBuffer.clear(userID)

	err := h.storage.SendMessage(chatID, "Создаю файл в формате PDF. Это может занять немного времени")
	if err != nil {
		return err
	}

	fileIDs := h.photoBuffer.get(userID)
	if len(fileIDs) == 0 {
		err = h.storage.SendMessage(chatID, "Нет фото в буфере")
		if err != nil {
			return err
		}
		return nil
	}

	dir := fmt.Sprintf("%s_%d", h.path, userID)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Printf("Ошибка создания директории: %v", err)
		err = h.storage.SendMessage(chatID, "Ошибка сервера")
		if err != nil {
			return err
		}
		return nil
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
		err = h.storage.SendMessage(chatID, "Ошибка при генерации PDF.")
		if err != nil {
			return err
		}
		os.RemoveAll(dir)
		return nil
	}

	f, err := os.Open(pdfPath)
	if err != nil {
		log.Printf("Ошибка открытия PDF: %v", err)
		err = h.storage.SendMessage(chatID, "Не удалось открыть PDF.")
		if err != nil {
			return err
		}
		os.RemoveAll(dir)
		return nil
	}
	defer f.Close()

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileReader{
		Name:   "photos.pdf",
		Reader: f,
	})
	for i := 0; i < 3; i++ {
		err := h.storage.SendDocument(doc)
		if err == nil {
			metrics.PdfCount.Inc()
			err = h.storage.SendMessage(chatID, "Твой PDF готов.")
			if err != nil {
				return err
			}
			return nil
		}
		tgErr, ok := err.(*tgbotapi.Error)
		if !ok {
			log.Printf("Ошибка отправки PDF: %v", err)
			err = h.storage.SendMessage(chatID, "Ошибка при отправке PDF.")
			if err != nil {
				return err
			}
			return nil
		}
		if tgErr.Code == 400 && tgErr.Message == "Bad Request: message text is empty" {
			continue
		}
		log.Printf("Ошибка отправки PDF: %v", err)
		err = h.storage.SendMessage(chatID, "Ошибка при отправке PDF.")
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (h *Handler) NewFile(chatID, userID int64) error {
	buf := h.photoBuffer.get(userID)
	if len(buf) > 0 {
		err := h.storage.SendMessage(chatID, "Ты еще не закончил с предыдущим файлом. Нажми \"Завершить\" или \"Очистить\"")
		if err != nil {
			return err
		}
		return nil
	}
	err := h.storage.SendMessage(chatID, "Отправь мне изображения для нового файла. После того, как отправишь все нужные фото, нажми \"Завершить\"")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Help(chatID int64) error {
	err := h.storage.SendMessage(chatID, "Если ты получил ошибку, нажми \"Новый файл\". \nЕсли хочешь перезапустить бот, отправь \"\\start\". \nЕсли хочешь удалить фото без создания файла, нажми \"Очистить\"")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Photo(update tgbotapi.Update, userID int64) {
	metrics.PhotoCount.Inc()
	sourcePhoto := update.Message.Photo[len(update.Message.Photo)-1]
	h.photoBuffer.add(userID, sourcePhoto.FileID)
}
