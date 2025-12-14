package port

import "github.com/andreychano/compressor-golang/internal/pkg/core/domain"

// Processor defines a contract for any compression algorithm
type Processor interface {
	Process(inputFile domain.File, opts domain.Options) (domain.File, error)
	Supports(mimeType string) bool
}
