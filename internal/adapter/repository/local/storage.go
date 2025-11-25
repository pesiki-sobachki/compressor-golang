package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings" // <--- Важный импорт для защиты

	"github.com/andreychano/compressor-golang/pkg/core/domain"
)

type LocalFileStorage struct {
	basePath string
}

func NewLocalFileStorage(basePath string) *LocalFileStorage {
	return &LocalFileStorage{
		basePath: basePath,
	}
}

// Save сохраняет файл на диск
func (s *LocalFileStorage) Save(ctx context.Context, file domain.File, relativePath string) (string, error) {
	fullPath := filepath.Join(s.basePath, relativePath)

	// 1. Создаем папку
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// 2. Создаем файл
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// 3. Перемотка
	if _, err := file.Content.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek file content: %w", err)
	}

	// 4. Запись
	if _, err := io.Copy(dst, file.Content); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	return fullPath, nil
}

// Get ищет файл на диске и возвращает его открытым (ЭТОГО МЕТОДА НЕ ХВАТАЛО)
func (s *LocalFileStorage) Get(ctx context.Context, relativePath string) (domain.File, error) {
	// 1. ЗАЩИТА: Очищаем путь
	cleanRelative := filepath.Clean(relativePath)

	// Если в пути есть ".." (попытка выйти назад) или он начинается с корня "/"
	if strings.Contains(cleanRelative, "..") || filepath.IsAbs(cleanRelative) {
		return domain.File{}, fmt.Errorf("access denied: invalid path")
	}

	// 2. Склеиваем с базовой папкой
	fullPath := filepath.Join(s.basePath, cleanRelative)

	// 3. Открываем файл (только для чтения)
	file, err := os.Open(fullPath)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to open file: %w", err)
	}

	// 4. Узнаем размер
	stat, err := file.Stat()
	if err != nil {
		file.Close() // Если не смогли узнать размер, закрываем файл
		return domain.File{}, fmt.Errorf("failed to get file info: %w", err)
	}

	// 5. Возвращаем. Файл остается ОТКРЫТЫМ.
	return domain.File{
		Content:  file,
		Size:     stat.Size(),
		MimeType: "application/octet-stream",
	}, nil
}
