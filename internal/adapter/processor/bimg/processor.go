package bimg

import (
	"bytes"
	"fmt"
	"github.com/andreychano/compressor-golang/core/domain"
	"github.com/h2non/bimg"
	"io"
)

type Processor struct {
}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) Supports(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func (p *Processor) Process(inputFile domain.File, opts domain.Options) (domain.File, error) {
	//Сброс чтения в начало
	if _, err := inputFile.Content.Seek(0, 0); err != nil {
		//неудачный сборс чтения
		return domain.File{}, fmt.Errorf("failed to seek file content: %w", err)
	}
	//Чтение и запись в буфер
	buffer, err := io.ReadAll(inputFile.Content)
	if err != nil {
		return domain.File{}, fmt.Errorf("Cannot read: %w", err)
	}

	/*Вернем заглушку

	return domain.File{}, nil
	*/

	img := bimg.NewImage(buffer) // Посыл файла в компрессор

	//Настройка параметров сжатия

	processOptions := bimg.Options{
		Quality:       opts.Quality, //Берем качество из наших настроек
		StripMetadata: true,         // Удаляем лишние метаданные
	}

	//сжатие

	processedBuffer, err := img.Process(processOptions)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to process image: %w", err)
	}

	// Теперь есть буфер processedBuffer со сжатыми байтами
	//Теперь надо вернуть их как domain.File

	return domain.File{
		//Превращаем байты обратно в reader
		Content: bytes.NewReader(processedBuffer),
		//укажем что это (посто напишем что было или jpeg)
		// для простоты пока просто пишем "image/"
		MimeType: "image/" + opts.Format,
		//размер новых данных
		Size: int64(len(processedBuffer)),
	}, nil

}
