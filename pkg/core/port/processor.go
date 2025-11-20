package port

import "github.com/andreychano/compressor-golang/pkg/core/domain"

// Processor определяет контракт для любого алгоритма компрессии.
type Processor interface {
	// Process применяет компрессию к файлу.
	Process(inputFile domain.File, opts domain.Options) (domain.File, error)
	// Supports проверяет, может ли этот процессор обработать данный MIME-тип.
	Supports(mimeType string) bool
}
