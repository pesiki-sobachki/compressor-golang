package http

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local/pathvalidator"
	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

const maxUploadSize = 32 << 20 // 32 MB

type Handler struct {
	svc *service.CompressionService
}

func NewHandler(svc *service.CompressionService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/upload", h.upload)
	mux.HandleFunc("/process", h.process)
	mux.HandleFunc("/file", h.getFile)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	dFile, dOptions, err := parseParamsStd(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if closer, ok := dFile.Content.(io.Closer); ok {
		defer closer.Close()
	}

	savedPath, err := h.svc.CompressAndSave(r.Context(), dFile, dOptions)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"success","compressed_path":"%s","message":"File saved successfully"}`, savedPath)
}

func (h *Handler) process(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	dFile, dOptions, err := parseParamsStd(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if closer, ok := dFile.Content.(io.Closer); ok {
		defer closer.Close()
	}

	resultFile, err := h.svc.Process(dFile, dOptions)
	if err != nil {
		log.Printf("Process failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("processed.%s", dOptions.Format)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", resultFile.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(resultFile.Size, 10))

	if _, err := io.Copy(w, resultFile.Content); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (h *Handler) getFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path query parameter is required", http.StatusBadRequest)
		return
	}

	fileInfo, err := h.svc.GetFile(r.Context(), path)
	if err != nil {
		var vErr *pathvalidator.ValidationError
		if errors.As(err, &vErr) {
			http.Error(w, vErr.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "File not found or access denied", http.StatusNotFound)
		return
	}
	if closer, ok := fileInfo.Content.(io.Closer); ok {
		defer closer.Close()
	}

	fileName := filepath.Base(path)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	io.Copy(w, fileInfo.Content)
}

func parseParamsStd(r *http.Request) (domain.File, domain.Options, error) {
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("FormFile error: %v", err)
		return domain.File{}, domain.Options{}, fmt.Errorf("no file uploaded: %w", err)
	}

	outputFormat := r.FormValue("format")
	if outputFormat == "" {
		outputFormat = "jpeg"
	}

	qualityStr := r.FormValue("quality")
	if qualityStr == "" {
		qualityStr = "80"
	}
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality == 0 {
		quality = 80
	}

	return domain.File{
			Content:  file,
			MimeType: header.Header.Get("Content-Type"),
			Size:     header.Size,
		},
		domain.Options{
			Format:  outputFormat,
			Quality: quality,
		},
		nil
}
