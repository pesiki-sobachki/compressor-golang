package port

import (
	"context"

	"github.com/andreychano/compressor-golang/internal/pkg/core/domain"
)

// File repository defines a contract for file storage

type FileRepository interface {
	Save(ctx context.Context, file domain.File, path string) (string, error)
	Get(ctx context.Context, path string) (domain.File, error)
}
