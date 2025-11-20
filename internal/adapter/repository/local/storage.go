package local

import (
	"context"
	"fmt"
	"github.com/andreychano/compressor-golang/pkg/core/domain"
	"io"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	basePath string
}

func NewLocalFileStorage(basePath string) *LocalFileStorage {
	return &LocalFileStorage{
		basePath: basePath,
	}
}

func (s *LocalFileStorage) Save(ctx context.Context, file domain.File, relativePath string) (string, error) {

	fullPath := filepath.Join(s.basePath, relativePath)

	//1. Получаем путь к папке (отрезаем имя файла)
	dir := filepath.Dir(fullPath)

	//2. Создаем папку (и всё родительское)
	// 0755 - это права доступа(rwxr-rx-x), стандарт для папок
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Если не вышло создать, выводим ошибку
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	//3. Создаем файл на диске
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %w", err)
	}
	defer dst.Close()

	//4. Сброс чтения в начало
	if _, err := file.Content.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek file content: %w", err)
	}

	//5. Копируем данные из памяти (file.Content) в файл на диске (dst)
	if _, err := io.Copy(dst, file.Content); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	return fullPath, nil
}
