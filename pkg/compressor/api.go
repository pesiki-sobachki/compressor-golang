package compressor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

// Compressor is a high-level façade for image compression.
type Compressor struct {
	svc *service.CompressionService
}

// New creates a new Compressor façade.
//
// Здесь ты сам решаешь, как собирать зависимости:
// либо принимаешь уже готовый service.CompressionService,
// либо конфиг (basePath, processor и т.п.) и собираешь внутри.
func New(svc *service.CompressionService) *Compressor {
	return &Compressor{svc: svc}
}

// Compress reads input, applies compression options and returns compressed bytes
// along with basic metadata.
func (c *Compressor) Compress(ctx context.Context, r io.Reader, opts OptionsDTO) ([]byte, ResultDTO, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		return nil, ResultDTO{}, fmt.Errorf("failed to read input: %w", err)
	}

	file := domain.File{
		Content:  bytes.NewReader(buf.Bytes()),
		MimeType: "", // при желании можешь принимать MIME снаружи и сюда прокидывать
		Size:     int64(buf.Len()),
	}

	domainOpts := domain.Options{
		Format:    opts.Format,
		Quality:   opts.Quality,
		MaxWidth:  opts.MaxWidth,
		MaxHeight: opts.MaxHeight,
	}

	outFile, err := c.svc.Process(file, domainOpts)
	if err != nil {
		return nil, ResultDTO{}, err
	}

	outBuf := &bytes.Buffer{}
	if _, err := io.Copy(outBuf, outFile.Content); err != nil {
		return nil, ResultDTO{}, fmt.Errorf("failed to read processed file: %w", err)
	}

	return outBuf.Bytes(), ResultDTO{
		MimeType: outFile.MimeType,
		Size:     outFile.Size,
	}, nil
}
