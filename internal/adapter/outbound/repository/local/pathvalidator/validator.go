package pathvalidator

import (
	"path/filepath"
	"strings"
)

// Validator performs strict filesystem path validation.
type Validator struct {
	basePath string
}

// New creates a new validator with the given base path.
func New(basePath string) *Validator {
	return &Validator{
		basePath: filepath.Clean(basePath),
	}
}

// Validate checks that the given relative path is safe to use.
func (v *Validator) Validate(relativePath string) error {
	if relativePath == "" {
		return NewValidationError("empty path")
	}

	cleanRelative := filepath.Clean(relativePath)

	// 1. Directory traversal.
	if strings.Contains(cleanRelative, "..") {
		return NewValidationError("path traversal attempt: contains ..")
	}

	// 2. Absolute paths are not allowed.
	if filepath.IsAbs(cleanRelative) {
		return NewValidationError("absolute path not allowed")
	}

	// 3. Null bytes.
	if strings.ContainsAny(cleanRelative, "\x00") {
		return NewValidationError("null bytes not allowed")
	}

	// 4. Ensure the resolved path does not escape basePath.
	resolvedPath := filepath.Join(v.basePath, cleanRelative)
	resolvedClean := filepath.Clean(resolvedPath)

	if !strings.HasPrefix(resolvedClean, v.basePath) {
		return NewValidationError("path escapes base directory")
	}

	// 5. Disallow unsafe characters (Windows/Unix).
	unsafeChars := `*|"<>?`
	for _, char := range unsafeChars {
		if strings.ContainsRune(cleanRelative, char) {
			return NewValidationError("unsafe characters not allowed")
		}
	}

	// 6. Maximum length.
	if len(cleanRelative) > 4096 {
		return NewValidationError("path too long")
	}

	// 7. Leading slashes are not allowed.
	if strings.HasPrefix(cleanRelative, "/") || strings.HasPrefix(cleanRelative, "\\") {
		return NewValidationError("leading slash not allowed")
	}

	return nil
}

// ValidationError is a typed error for path validation failures.
type ValidationError struct {
	Reason string
}

func NewValidationError(reason string) *ValidationError {
	return &ValidationError{Reason: reason}
}

func (e *ValidationError) Error() string {
	return e.Reason
}
