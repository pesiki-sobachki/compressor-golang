package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/andreychano/compressor-golang/internal/adapter/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/repository/local"
	"github.com/andreychano/compressor-golang/pkg/core/domain"
	"github.com/andreychano/compressor-golang/pkg/core/service"
)

func main() {
	// 1. Инициализация
	storage := local.NewLocalFileStorage("storage")
	processor := bimg.NewProcessor()
	svc := service.NewCompressionService(storage, processor)

	// 2. Настройка сервера
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20 // 10 MB

	// --- РОУТ 1: Сохранить и Сжать ---
	r.POST("/upload", func(c *gin.Context) {
		dFile, dOptions, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if closer, ok := dFile.Content.(io.Closer); ok {
			defer closer.Close()
		}

		savedPath, err := svc.CompressAndSave(c.Request.Context(), dFile, dOptions)
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
	})

	// --- РОУТ 2: Только Сжать (Без сохранения) ---
	r.POST("/process", func(c *gin.Context) {
		dFile, dOptions, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if closer, ok := dFile.Content.(io.Closer); ok {
			defer closer.Close()
		}

		resultFile, err := svc.Process(dFile, dOptions)
		if err != nil {
			log.Printf("Process failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Отдаем результат сразу (скачивание)
		filename := fmt.Sprintf("processed.%s", dOptions.Format)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", resultFile.MimeType)
		c.Header("Content-Length", strconv.FormatInt(resultFile.Size, 10))

		if _, err := io.Copy(c.Writer, resultFile.Content); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	})

	// --- РОУТ 3: Скачать файл ---
	r.GET("/file", func(c *gin.Context) {
		path := c.Query("path")
		if path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
			return
		}

		fileInfo, err := svc.GetFile(c.Request.Context(), path)
		if err != nil {
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
	})

	// Запуск
	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// Вспомогательная функция
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
