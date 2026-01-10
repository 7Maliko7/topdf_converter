package bot

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/signintech/gopdf"
)

func downloadPhoto(url, savePath string) error {
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
		if err := pdf.Image(img, 0, 0, gopdf.PageSizeA4); err != nil {
			return fmt.Errorf("ошибка вставки изображения %s: %w", img, err)
		}
	}

	if err := pdf.WritePdf(outputPath); err != nil {
		return fmt.Errorf("WritePdf error: %w", err)
	}
	return nil
}
