package main

import (
	"ddownload/internal/handlers"
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handlers.HandleIndex)
	http.HandleFunc("/files", handlers.HandleFilesTable)
	http.HandleFunc("/folder", handlers.HandleFolderContent)
	http.HandleFunc("/download", handlers.HandleDownload)

	http.HandleFunc("/health", handlers.HandleHealth)

	fmt.Printf("http://localhost:%s\n", "3000")
	http.ListenAndServe(":"+"3000", nil)
}
