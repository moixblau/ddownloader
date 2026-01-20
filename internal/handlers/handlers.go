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
	"runtime"
	"runtime/debug"
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

	tmpl := template.Must(template.ParseFiles("templates/table.html"))
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
		runtime.GC()
		debug.FreeOSMemory()
		return
	} else {
		file, _ := os.Open(path)
		defer file.Close()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	}
}
