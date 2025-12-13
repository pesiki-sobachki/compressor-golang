package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andreychano/compressor-golang/internal/adapter/repository/local/pathvalidator"
	"github.com/andreychano/compressor-golang/pkg/core/domain"
)

type LocalFileStorage struct {
	basePath      string
	pathValidator *pathvalidator.Validator
}

// NewLocalFileStorage инициализирует сторедж и валидатор путей.
func NewLocalFileStorage(basePath string) *LocalFileStorage {
	return &LocalFileStorage{
		basePath:      basePath,
		pathValidator: pathvalidator.New(basePath),
	}
}

func (s *LocalFileStorage) Save(ctx context.Context, file domain.File, relativePath string) (string, error) {
	// 1. Валидация относительного пути
	if err := s.pathValidator.Validate(relativePath); err != nil {
		// TODO: здесь потом WARN-лог вида:
		// logger.Warnf("path validation failed in Save: %v", err)
		return "", fmt.Errorf("access denied: invalid path")
	}

	// 2. Строим абсолютный путь только после успешной валидации
	fullPath := filepath.Join(s.basePath, relativePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := file.Content.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek file content: %w", err)
	}

	if _, err := io.Copy(dst, file.Content); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	return fullPath, nil
}

func (s *LocalFileStorage) Get(ctx context.Context, relativePath string) (domain.File, error) {
	// было:
	// if err := s.pathValidator.Validate(relativePath); err != nil {
	//     return domain.File{}, fmt.Errorf("access denied: invalid path")
	// }

	// нужно так:
	if err := s.pathValidator.Validate(relativePath); err != nil {
		// не меняем тип ошибки, просто пробрасываем
		return domain.File{}, err
	}

	fullPath := filepath.Join(s.basePath, relativePath)

	f, err := os.Open(fullPath)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return domain.File{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return domain.File{
		Content:  f,
		Size:     stat.Size(),
		MimeType: "application/octet-stream",
	}, nil
}
