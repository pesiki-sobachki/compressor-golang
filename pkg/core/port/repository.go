package port

import (
	"context"

	"github.com/andreychano/compressor-golang/pkg/core/domain"
)

// File repisitory определяет контракт для хранения файлов

type FileRepository interface {
	Save(ctx context.Context, file domain.File, path string) (string, error)
}
