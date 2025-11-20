package service

import (
	"context"
	"fmt"
	"path/filepath" // Для работы с путями

	"github.com/andreychano/compressor-golang/core/domain"
	"github.com/andreychano/compressor-golang/core/port"
	"github.com/google/uuid" // для генерации уникальных имен
)

// CompressionService оркестрирует процесс компрессии.
type CompressionService struct {
	processors []port.Processor
	repository port.FileRepository
}

// NewCompressionService создает новый сервис. Мы передаем ему реализации портов.
func NewCompressionService(repo port.FileRepository, processors ...port.Processor) *CompressionService {
	return &CompressionService{
		repository: repo,
		processors: processors,
	}
}

// CompressFile - основной метод, выполняющий всю работу.
func (s *CompressionService) CompressFile(ctx context.Context, file domain.File, opts domain.Options) (string, error) {
	var selectedProcessor port.Processor
	// 1. Найти подходящий процессор
	for _, p := range s.processors {
		if p.Supports(file.MimeType) {
			selectedProcessor = p
			break
		}
	}

	if selectedProcessor == nil {
		return "", fmt.Errorf("unsupported file type: %s", file.MimeType)
	}

	// 2. Обработать файл
	compressedFile, err := selectedProcessor.Process(file, opts)
	if err != nil {
		return "", fmt.Errorf("failed to process file: %w", err)
	}

	// 3. Сохранить файл
	// Генерируем уникальное имя, чтобы файлы не перезаписывали друг друга
	uniqueID := uuid.New().String()
	// Используем формат из опций для расширения файла
	fileName := fmt.Sprintf("%s.%s", uniqueID, opts.Format)
	filePath := filepath.Join("compressed", fileName)

	savedPath, err := s.repository.Save(ctx, compressedFile, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return savedPath, nil
}
