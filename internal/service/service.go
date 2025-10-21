package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg" // Для декодирования JPEG
	_ "image/png"  // Для декодирования PNG
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/image/webp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/signintech/gopdf"
)

func SaveMediaGroup(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) ([]string, error) {
	mediaGroupBuffers := make(map[string][]*tgbotapi.Message)
	bufferLock := sync.Mutex{}

	bufferLock.Lock()
	defer bufferLock.Unlock()

	groupID := msg.MediaGroupID
	mediaGroupBuffers[groupID] = append(mediaGroupBuffers[groupID], msg)
	ch := make(chan []string)
	chErr := make(chan error)
	go func(id string) {
		time.Sleep(1 * time.Second)
		bufferLock.Lock()
		messages := mediaGroupBuffers[id]
		delete(mediaGroupBuffers, id)
		list, err := processMediaGroup(bot, messages)
		if err != nil {
			chErr <- err
		}
		ch <- list

	}(groupID)

	result := <-ch
	return result, nil
}

func processMediaGroup(bot *tgbotapi.BotAPI, messages []*tgbotapi.Message) ([]string, error) {
	if len(messages) == 0 {
		return nil, errors.New("no messages")
	}
	pathList := make([]string, len(messages))
	for _, msg := range messages {
		path, err := UploadPhoto(bot, msg)
		if err != nil {
			return nil, err
		}
		pathList = append(pathList, path)
	}
	return pathList, nil
}

func UploadPhoto(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) (string, error) {
	var fileID string
	if msg.Photo != nil {
		fileID = (msg.Photo)[len(msg.Photo)-1].FileID
	} else if msg.Document != nil {
		fileID = msg.Document.FileID
	}
	path, err := UploadImage(bot, msg.Chat.ID, int64(msg.Contact.UserID), fileID)
	if err != nil {
		return "", err
	}
	return path, nil
}

// Функция для обработки изображения и отправки PDF
func UploadImage(bot *tgbotapi.BotAPI, chatID, userID int64, fileID string) (string, error) {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", err
	}

	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", err
	}
	img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		img, err = webp.Decode(bytes.NewReader(buf.Bytes()))
		if err != nil {
			return "", err
		}
	}

	userDir := filepath.Join("downloads", fmt.Sprintf("%d", userID))
	err = os.MkdirAll(userDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(userDir, fmt.Sprintf("%s.jpg", file.FileID))
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	defer out.Close()

	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", err
	}

	filePath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func MakePDF(imgPaths []string, pdfPath string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	for _, img := range imgPaths {
		pdf.AddPage()
		err := pdf.Image(img, 0, 0, gopdf.PageSizeA4)
		if err != nil {
			return err
		}
	}
	output := fmt.Sprintf("%s.pdf", pdfPath)
	return pdf.WritePdf(output)
}

func SendPDF(bot *tgbotapi.BotAPI, chatID int64, pdfPath string) error {
	msg := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(pdfPath))
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}
	return err
}
