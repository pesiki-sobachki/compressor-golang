package main

import (
	"context"
	"log"
	"net/http"

	httpadapter "github.com/andreychano/compressor-golang/internal/adapter/inbound/http"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local"
	"github.com/andreychano/compressor-golang/internal/config"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad(ctx)

	storage := local.NewLocalFileStorage(cfg.Storage.Path)
	processor := bimg.NewProcessor()
	svc := service.NewCompressionService(storage, processor)

	mux := http.NewServeMux()
	h := httpadapter.NewHandler(svc)
	h.RegisterRoutes(mux)

	log.Printf("Starting server on %s...", cfg.HTTP.Address)
	if err := http.ListenAndServe(cfg.HTTP.Address, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
