package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func (s *LocalFileStorage) Save(ctx context.Context, file domain.File, relativePath string) (string, error) {
	fullPath := filepath.Join(s.basePath, relativePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
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

	cleanRelative := filepath.Clean(relativePath)

	if strings.Contains(cleanRelative, "..") || filepath.IsAbs(cleanRelative) {
		return domain.File{}, fmt.Errorf("access denied: invalid path")
	}

	fullPath := filepath.Join(s.basePath, cleanRelative)

	file, err := os.Open(fullPath)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return domain.File{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return domain.File{
		Content:  file,
		Size:     stat.Size(),
		MimeType: "application/octet-stream",
	}, nil
}
