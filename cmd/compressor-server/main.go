package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	//Локальные импорты
	"github.com/andreychano/compressor-golang/internal/adapter/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/repository/local"
	"github.com/andreychano/compressor-golang/pkg/core/domain"
	"github.com/andreychano/compressor-golang/pkg/core/service"
)

func main() {
	//1. Инициализация адаптеров (Инфраструктура)
	storage := local.NewLocalFileStorage("storage")
	processor := bimg.NewProcessor()

	//2. Инициализация сервиса
	svc := service.NewCompressionService(storage, processor)

	//3. Настройка HTTP-сервера
	r := gin.Default()
	r.MaxMultipartMemory = 10 << 20

	r.POST("/upload", func(c *gin.Context) {
		// 1. Получаем файл
		fileHeader, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer file.Close()

		// 2. Читаем параметры
		outputFormat := c.DefaultPostForm("format", "jpeg")
		qualityStr := c.DefaultPostForm("quality", "80")

		quality, err := strconv.Atoi(qualityStr)
		if err != nil {
			quality = 80
		}

		// 3. Готовим данные
		domainFile := domain.File{
			Content:  file,
			MimeType: fileHeader.Header.Get("Content-Type"),
			Size:     fileHeader.Size,
		}

		domainOptions := domain.Options{
			Format:  outputFormat,
			Quality: quality,
		}

		// 4. Вызываем сервис
		savedPath, err := svc.CompressFile(c.Request.Context(), domainFile, domainOptions)
		if err != nil {
			log.Printf("Compression failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 5. Успех
		c.JSON(http.StatusOK, gin.H{
			"status":          "success",
			"original_size":   fileHeader.Size,
			"compressed_path": savedPath,
			"message":         "File compressed and saved successfully!",
		})
	})

	//4. Запуск сервера
	log.Println("Starting server on: 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
