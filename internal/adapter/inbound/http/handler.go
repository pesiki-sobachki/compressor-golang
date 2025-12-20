package http

import (
	"errors"
	"fmt"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local/pathvalidator"
	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/andreychano/compressor-golang/internal/core/service"
	applogger "github.com/andreychano/compressor-golang/internal/logger"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

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

	applogger.Log.Info().Msg("upload handler called")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dFile, dOptions, err := parseParamsStd(w, r)
	if err != nil {
		return
	}
	if closer, ok := dFile.Content.(io.Closer); ok {
		defer closer.Close()
	}

	savedPath, err := h.svc.CompressAndSave(r.Context(), dFile, dOptions)
	if err != nil {
		applogger.Log.Error().
			Err(err).
			Msg("upload failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	const mib = 1024 * 1024
	sizeMiB := float64(dFile.Size) / float64(mib)
	sizeMiBStr := fmt.Sprintf("%.2f", sizeMiB)

	applogger.Log.Info().
		Str("path", savedPath).
		Int("quality", dOptions.Quality).
		Str("format", dOptions.Format).
		//Int64("size_bytes", dFile.Size).
		Str("size_mib", sizeMiBStr).
		Str("remote_addr", r.RemoteAddr).
		Msg("upload succeeded")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"success","compressed_path":"%s","message":"File saved successfully"}`, savedPath)
}

func (h *Handler) process(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dFile, dOptions, err := parseParamsStd(w, r)
	if err != nil {
		return
	}
	if closer, ok := dFile.Content.(io.Closer); ok {
		defer closer.Close()
	}

	resultFile, err := h.svc.Process(dFile, dOptions)
	if err != nil {
		applogger.Log.Error().
			Err(err).
			Msg("process failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	applogger.Log.Info().
		Int("quality", dOptions.Quality).
		Str("format", dOptions.Format).
		Int64("input_size", dFile.Size).
		Int64("output_size", resultFile.Size).
		Str("remote_addr", r.RemoteAddr).
		Msg("process succeeded")

	filename := fmt.Sprintf("processed.%s", dOptions.Format)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", resultFile.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(resultFile.Size, 10))

	if _, err := io.Copy(w, resultFile.Content); err != nil {
		applogger.Log.Error().
			Err(err).
			Msg("failed to write response")
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

	applogger.Log.Info().
		Str("path", path).
		Int64("size", fileInfo.Size).
		Str("remote_addr", r.RemoteAddr).
		Msg("get file succeeded")

	fileName := filepath.Base(path)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	io.Copy(w, fileInfo.Content)
}

func parseParamsStd(w http.ResponseWriter, r *http.Request) (domain.File, domain.Options, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			applogger.Log.Warn().
				Err(err).
				Msg("request body too large")
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			return domain.File{}, domain.Options{}, fmt.Errorf("request body too large: %w", err)
		}

		applogger.Log.Error().
			Err(err).
			Msg("FormFile error")
		http.Error(w, "no file uploaded", http.StatusBadRequest)
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
