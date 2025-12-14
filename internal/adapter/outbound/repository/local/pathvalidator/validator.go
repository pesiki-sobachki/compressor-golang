//go:build !test

package pathvalidator

import (
	"path/filepath"
	"strings"
)

// Validator для строгой проверки путей файловой системы
type Validator struct {
	basePath string
}

// New создает новый валидатор с базовым путем
func New(basePath string) *Validator {
	return &Validator{
		basePath: filepath.Clean(basePath),
	}
}

// Validate проверяет путь на безопасность
func (v *Validator) Validate(relativePath string) error {
	if relativePath == "" {
		return NewValidationError("empty path")
	}

	cleanRelative := filepath.Clean(relativePath)

	// 1. Directory traversal
	if strings.Contains(cleanRelative, "..") {
		return NewValidationError("path traversal attempt: contains ..")
	}

	// 2. Абсолютный путь
	if filepath.IsAbs(cleanRelative) {
		return NewValidationError("absolute path not allowed")
	}

	// 3. Null bytes
	if strings.ContainsAny(cleanRelative, "\x00") {
		return NewValidationError("null bytes not allowed")
	}

	// 4. Выход за пределы basePath (самый надежный чек)
	resolvedPath := filepath.Join(v.basePath, cleanRelative)
	resolvedClean := filepath.Clean(resolvedPath)

	if !strings.HasPrefix(resolvedClean, v.basePath) {
		return NewValidationError("path escapes base directory")
	}

	// 5. Unsafe символы (Windows/Unix)
	unsafeChars := `*|"<>?`
	for _, char := range unsafeChars {
		if strings.ContainsRune(cleanRelative, rune(char)) {
			return NewValidationError("unsafe characters not allowed")
		}
	}

	// 6. Максимальная длина
	if len(cleanRelative) > 4096 {
		return NewValidationError("path too long")
	}

	// 7. Leading slashes
	if strings.HasPrefix(cleanRelative, "/") || strings.HasPrefix(cleanRelative, "\\") {
		return NewValidationError("leading slash not allowed")
	}

	return nil
}

// ValidationError типизированная ошибка
type ValidationError struct {
	Reason string
}

func NewValidationError(reason string) *ValidationError {
	return &ValidationError{Reason: reason}
}

func (e *ValidationError) Error() string {
	return e.Reason
}
