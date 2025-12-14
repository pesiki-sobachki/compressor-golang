package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/andreychano/compressor-golang/internal/core/port"
	"github.com/google/uuid"
)

type CompressionService struct {
	processors []port.Processor
	repository port.FileRepository
}

func NewCompressionService(repo port.FileRepository, processors ...port.Processor) *CompressionService {
	return &CompressionService{
		repository: repo,
		processors: processors,
	}
}

func (s *CompressionService) Process(file domain.File, opts domain.Options) (domain.File, error) {
	var selectedProcessor port.Processor

	for _, p := range s.processors {
		if p.Supports(file.MimeType) {
			selectedProcessor = p
			break
		}
	}

	if selectedProcessor == nil {
		return domain.File{}, fmt.Errorf("unsupport file type: %s", file.MimeType)
	}

	return selectedProcessor.Process(file, opts)
}

func (s *CompressionService) CompressAndSave(ctx context.Context, file domain.File, opts domain.Options) (string, error) {

	compressedFile, err := s.Process(file, opts)
	if err != nil {
		return "", err
	}

	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s.%s", uniqueID, opts.Format)

	filePath := filepath.Join("compressed", fileName)

	return s.repository.Save(ctx, compressedFile, filePath)
}

func (s *CompressionService) GetFile(ctx context.Context, path string) (domain.File, error) {
	return s.repository.Get(ctx, path)
}
