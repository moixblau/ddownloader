package handlers

import (
	"archive/zip"
	"ddownload/internal/models"
	"ddownload/internal/service"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func HandleFilesTable(w http.ResponseWriter, r *http.Request) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

	nodes, err := service.ScanDirectory(dataDir, 0)
	if err != nil {
		if os.IsNotExist(err) {
			nodes = []models.FileNode{}
		} else {
			http.Error(w, "Error scanning directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for i := range nodes {
		service.FillFormatSize(&nodes[i])
	}

	funcMap := template.FuncMap{
		"multiply": func(a int, b float64) float64 {
			return float64(a) * b
		},
	}

	tmpl := template.Must(template.New("table.html").Funcs(funcMap).ParseFiles("templates/table.html", "templates/rows.html"))
	tmpl.Execute(w, nodes)
}

func HandleFolderContent(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	levelStr := r.URL.Query().Get("level")
	level := 0
	if levelStr != "" {
		fmt.Sscanf(levelStr, "%d", &level)
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	rel, err := filepath.Rel(absDataDir, absPath)
	if err != nil || (len(rel) >= 2 && rel[:2] == "..") {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	nodes, err := service.ScanDirectory(path, level+1)
	if err != nil {
		http.Error(w, "Error scanning directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range nodes {
		service.FillFormatSize(&nodes[i])
	}

	funcMap := template.FuncMap{
		"multiply": func(a int, b float64) float64 {
			return float64(a) * b
		},
	}

	tmpl := template.Must(template.New("rows.html").Funcs(funcMap).ParseFiles("templates/rows.html"))
	tmpl.Execute(w, nodes)
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	stat, err := os.Stat(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if stat.IsDir() {
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", stat.Name()))

		zw := zip.NewWriter(w)
		defer zw.Close()

		ctx := r.Context()

		err := filepath.WalkDir(path, func(fpath string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if d.IsDir() {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(path, fpath)
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = relPath
			header.Method = zip.Store

			zipFile, err := zw.CreateHeader(header)
			if err != nil {
				return err
			}

			fsFile, err := os.Open(fpath)
			if err != nil {
				return err
			}
			defer fsFile.Close()

			_, err = io.Copy(zipFile, fsFile)
			return err
		})

		if err != nil {
			fmt.Printf("Error during download: %v\n", err)
		}
		return
	} else {
		file, _ := os.Open(path)
		defer file.Close()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	}
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
