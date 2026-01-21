package main

import (
	"ddownload/internal/config"
	"ddownload/internal/handlers"
	"ddownload/internal/service"
	"fmt"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()
	fileService := service.NewFileService(cfg.DataDir)
	h := handlers.NewHandler(cfg, fileService)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handlers.LoggingMiddleware(cfg.Logger, h.HandleIndex))
	http.HandleFunc("/files", handlers.LoggingMiddleware(cfg.Logger, h.HandleFilesTable))
	http.HandleFunc("/folder", handlers.LoggingMiddleware(cfg.Logger, h.HandleFolderContent))
	http.HandleFunc("/download", handlers.LoggingMiddleware(cfg.Logger, h.HandleDownload))
	http.HandleFunc("/delete", handlers.LoggingMiddleware(cfg.Logger, h.HandleDelete))

	http.HandleFunc("/health", h.HandleHealth)

	fmt.Printf("Starting server on http://localhost:%s\n", cfg.Port)
	cfg.Logger.Info("Server started", "port", cfg.Port, "dataDir", cfg.DataDir)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		cfg.Logger.Error("Server failed", "error", err)
	}
}
