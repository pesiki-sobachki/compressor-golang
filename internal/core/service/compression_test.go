package service_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/andreychano/compressor-golang/internal/core/domain"
	portmocks "github.com/andreychano/compressor-golang/internal/core/port/mocks"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

func TestCompressionService_CompressAndSave_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := portmocks.NewMockFileRepository(ctrl)
	processorMock := portmocks.NewMockProcessor(ctrl)

	s := service.NewCompressionService(repoMock, processorMock)

	file := domain.File{MimeType: "image/png"}
	opts := domain.Options{Format: "jpeg"}

	processorMock.EXPECT().
		Supports(file.MimeType).
		Return(true)

	compressed := domain.File{MimeType: "image/jpeg"}

	processorMock.EXPECT().
		Process(file, opts).
		Return(compressed, nil)

	repoMock.EXPECT().
		Save(gomock.Any(), compressed, gomock.Any()).
		Return("compressed/some-id.jpeg", nil)

	path, err := s.CompressAndSave(context.Background(), file, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatalf("expected non-empty path")
	}
}
