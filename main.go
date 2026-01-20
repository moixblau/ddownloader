package main

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileNode struct {
	Name       string
	Path       string
	TotalSize  int64
	FormatSize string
	IsFolder   bool
	Icon       string
	Color      string
	Children   []FileNode
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/files", handleFilesTable)
	http.HandleFunc("/download", handleDownload)

	fmt.Printf("http://localhost:%s\n", "3000")
	http.ListenAndServe(":"+"3000", nil)
}

func handleFilesTable(w http.ResponseWriter, r *http.Request) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

	nodes, err := scanDirectory(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			nodes = []FileNode{}
		} else {
			http.Error(w, "Error scanning directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for i := range nodes {
		fillFormatSize(&nodes[i])
	}

	tmpl := template.Must(template.ParseFiles("templates/table.html"))
	tmpl.Execute(w, nodes)
}

func fillFormatSize(node *FileNode) {
	node.FormatSize = formatBytes(node.TotalSize)

	for i := range node.Children {
		fillFormatSize(&node.Children[i])
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
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

		filepath.Walk(path, func(fpath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			return func() error {
				relPath, _ := filepath.Rel(path, fpath)
				zipFile, err := zw.Create(relPath)
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
			}()
		})
	} else {
		file, _ := os.Open(path)
		defer file.Close()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	}
}

func scanDirectory(rootPath string) ([]FileNode, error) {
	var nodes []FileNode

	files, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file.Name())
		info, _ := file.Info()

		icon, color := getFileMeta(file.Name(), file.IsDir())

		node := FileNode{
			Name:     file.Name(),
			Path:     fullPath,
			IsFolder: file.IsDir(),
			Icon:     icon,
			Color:    color,
		}

		if file.IsDir() {
			childScanner, _ := scanDirectory(fullPath)
			for _, child := range childScanner {
				node.TotalSize += child.TotalSize
			}
			node.Children = childScanner
		} else {
			node.TotalSize = info.Size()
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func getFileMeta(path string, isDir bool) (string, string) {
	if isDir {
		return "pi-folder", "#fbc02d"
	}
	ext := filepath.Ext(path)
	switch ext {
	case ".mp4", ".mkv", ".avi":
		return "pi-video", "#ff5722"
	case ".jpg", ".png", ".gif":
		return "pi-image", "#4caf50"
	case ".pdf":
		return "pi-file-pdf", "#f44336"
	case ".zip", ".rar":
		return "pi-file-archive", "#9c27b0"
	default:
		return "pi-file", "var(--primary-color)"
	}
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(1024), 0
	for n := b / 1024; n >= 1024; n /= 1024 {
		div *= 1024
		exp++
	}
	unit := []string{"KB", "MB", "GB", "TB"}[exp]
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), unit)
}
