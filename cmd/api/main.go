package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/andreychano/compressor-golang/internal/adapter/inbound/http"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local"
	"github.com/andreychano/compressor-golang/internal/pkg/core/service"
)

func main() {
	storage := local.NewLocalFileStorage("storage")
	processor := bimg.NewProcessor()
	svc := service.NewCompressionService(storage, processor)

	r := gin.Default()

	h := http.NewHandler(svc)
	h.RegisterRoutes(r)

	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
