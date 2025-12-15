package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shanth1/gotools/log"

	"github.com/andreychano/compressor-golang/internal/adapter/inbound/http"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

func main() {
	storage := local.NewLocalFileStorage("storage")
	processor := bimg.NewProcessor()
	svc := service.NewCompressionService(storage, processor)

	r := gin.Default()

	logger := log.NewFromConfig(log.Config{
		Level:        "debug",
		App:          "app_name",
		Service:      "service_name",
		UDPAddress:   ":5555",
		EnableCaller: false,
		Console:      true,
		JSONOutput:   false,
	})

	h := http.NewHandler(svc)
	h.RegisterRoutes(r)

	logger.Info().Str("port", ":8080").Msg("hello")
	logger.Warn().Msgf("hello from: %s", "warn")

	if err := r.Run(":8080"); err != nil {
		logger.Fatal().Msgf("failed to start server: %v", err)
	}
}
