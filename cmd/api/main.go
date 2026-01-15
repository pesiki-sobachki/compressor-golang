package main

import (
	"context"
	"fmt"
	"net/http"

	httpadapter "github.com/andreychano/compressor-golang/internal/adapter/inbound/http"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local"
	"github.com/andreychano/compressor-golang/internal/config"
	"github.com/andreychano/compressor-golang/internal/core/service"
	applogger "github.com/andreychano/compressor-golang/internal/logger"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad(ctx)

	// --- ОТЛАДКА (ВРЕМЕННО) ---
	// Это покажет нам, что реально загрузилось
	fmt.Printf("DEBUG: Config Address: %s\n", cfg.HTTP.Address)
	fmt.Printf("DEBUG: Logger Level: %s\n", cfg.Log.Level)
	fmt.Printf("DEBUG: Logger UDP: '%s'\n", cfg.Log.UDPAddress)
	// ---------------------------

	applogger.Init(cfg.Log)

	storage := local.NewLocalFileStorage(cfg.Storage.Path)
	processor := bimg.NewProcessor()
	svc := service.NewCompressionService(storage, *cfg, processor)

	mux := http.NewServeMux()
	h := httpadapter.NewHandler(svc)
	h.RegisterRoutes(mux)

	maxBytes := cfg.HTTP.MaxUploadSizeBytes()
	handler := httpadapter.MaxUploadSize(maxBytes, mux)

	applogger.Log.Info().
		Str("address", cfg.HTTP.Address).
		Int64("max_bytes", maxBytes).
		Msg("Starting server")

	if err := http.ListenAndServe(cfg.HTTP.Address, handler); err != nil {
		applogger.Log.Error().
			Err(err).
			Msg("failed to start server")
	}
}
