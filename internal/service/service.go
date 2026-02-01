package service

import (
	"fmt"
	"image"
	"io"
	"math"
	"net/http"
	"os"

	"github.com/signintech/gopdf"
)

func DownloadPhoto(url, savePath string) error {
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

func CreatePDFWithGoPDF(images []string, outputPath string) error {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	for _, img := range images {
		pdf.AddPage()
		pageSize, x, y, err := reshape(img)
		if err != nil {
			return err
		}
		if err := pdf.Image(img, x, y, pageSize); err != nil {
			return fmt.Errorf("ошибка вставки изображения %s: %w", img, err)
		}
	}

	if err := pdf.WritePdf(outputPath); err != nil {
		return fmt.Errorf("WritePdf error: %w", err)
	}
	return nil
}

func reshape(filename string) (*gopdf.Rect, float64, float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	imgCfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, 0, 0, err
	}

	origW := float64(imgCfg.Width)
	origH := float64(imgCfg.Height)

	pageW := gopdf.PageSizeA4.W
	pageH := gopdf.PageSizeA4.H

	margin := 40.0

	maxW := pageW - margin*2
	maxH := pageH - margin*2

	scale := math.Min(maxW/origW, maxH/origH)

	w := origW * scale
	h := origH * scale

	x := (pageW - w) / 2
	y := (pageH - h) / 2

	return &gopdf.Rect{
		W: w,
		H: h,
	}, x, y, nil
}
