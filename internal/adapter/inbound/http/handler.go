package http

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local/pathvalidator"
	"github.com/andreychano/compressor-golang/internal/pkg/core/domain"
	"github.com/andreychano/compressor-golang/internal/pkg/core/service"
)

type Handler struct {
	svc *service.CompressionService
}

func NewHandler(svc *service.CompressionService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.MaxMultipartMemory = 10 << 20

	r.POST("/upload", h.upload)
	r.POST("/process", h.process)
	r.GET("/file", h.getFile)
}

func (h *Handler) upload(c *gin.Context) {
	dFile, dOptions, err := parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if closer, ok := dFile.Content.(io.Closer); ok {
		defer closer.Close()
	}

	savedPath, err := h.svc.CompressAndSave(c.Request.Context(), dFile, dOptions)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"compressed_path": savedPath,
		"message":         "File saved successfully",
	})
}

func (h *Handler) process(c *gin.Context) {
	dFile, dOptions, err := parseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if closer, ok := dFile.Content.(io.Closer); ok {
		defer closer.Close()
	}

	resultFile, err := h.svc.Process(dFile, dOptions)
	if err != nil {
		log.Printf("Process failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("processed.%s", dOptions.Format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", resultFile.MimeType)
	c.Header("Content-Length", strconv.FormatInt(resultFile.Size, 10))

	if _, err := io.Copy(c.Writer, resultFile.Content); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (h *Handler) getFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
		return
	}

	fileInfo, err := h.svc.GetFile(c.Request.Context(), path)
	if err != nil {
		var vErr *pathvalidator.ValidationError
		if errors.As(err, &vErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": vErr.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or access denied"})
		return
	}
	if closer, ok := fileInfo.Content.(io.Closer); ok {
		defer closer.Close()
	}

	fileName := filepath.Base(path)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size, 10))

	io.Copy(c.Writer, fileInfo.Content)
}

func parseParams(c *gin.Context) (domain.File, domain.Options, error) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return domain.File{}, domain.Options{}, fmt.Errorf("no file uploaded")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return domain.File{}, domain.Options{}, fmt.Errorf("failed to open file")
	}

	outputFormat := c.DefaultPostForm("format", "jpeg")
	qualityStr := c.DefaultPostForm("quality", "80")
	quality, _ := strconv.Atoi(qualityStr)
	if quality == 0 {
		quality = 80
	}

	return domain.File{
		Content:  file,
		MimeType: fileHeader.Header.Get("Content-Type"),
		Size:     fileHeader.Size,
	}, domain.Options{Format: outputFormat, Quality: quality}, nil
}
