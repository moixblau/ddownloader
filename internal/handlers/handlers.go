package handlers

import (
	"archive/zip"
	"ddownload/internal/config"
	"ddownload/internal/models"
	"ddownload/internal/service"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type Handler struct {
	cfg                 *config.Config
	fileService         *service.FileService
	transmissionService *service.TransmissionService
	templates           *template.Template
}

func NewHandler(cfg *config.Config, fs *service.FileService) *Handler {
	funcMap := template.FuncMap{
		"multiply": func(a int, b float64) float64 {
			return float64(a) * b
		},
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"templates/index.html",
		"templates/table.html",
		"templates/rows.html",
		"templates/login.html",
	))

	ts := service.NewTransmissionService(cfg.TransmissionHost, cfg.TransmissionUser, cfg.TransmissionPass)

	return &Handler{
		cfg:                 cfg,
		fileService:         fs,
		transmissionService: ts,
		templates:           tmpl,
	}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Theme":       h.cfg.Theme,
		"AuthEnabled": h.cfg.IsAuthEnabled(),
	}); err != nil {
		h.cfg.Logger.Error("Error executing index template", "error", err)
	}
}

func (h *Handler) HandleFilesTable(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	var nodes []models.FileNode
	var err error

	if search != "" {
		nodes, err = h.fileService.SearchFiles(h.cfg.DataDir, search)
	} else {
		nodes, err = h.fileService.ScanDirectory(h.cfg.DataDir, 0)
	}

	if err != nil {
		if os.IsNotExist(err) {
			nodes = []models.FileNode{}
		} else {
			h.cfg.Logger.Error("Error getting files", "path", h.cfg.DataDir, "search", search, "error", err)
			http.Error(w, "Error scanning directory", http.StatusInternalServerError)
			return
		}
	}

	for i := range nodes {
		if nodes[i].FormatSize == "" {
			h.fileService.FillFormatSize(&nodes[i])
		}
	}

	if err := h.templates.ExecuteTemplate(w, "table.html", map[string]interface{}{
		"Nodes":               nodes,
		"TransmissionEnabled": h.cfg.IsTransmissionEnabled(),
	}); err != nil {
		h.cfg.Logger.Error("Error executing table template", "error", err)
	}
}

func (h *Handler) HandleActiveDownloads(w http.ResponseWriter, r *http.Request) {
	activeDownloads, err := h.transmissionService.GetActiveDownloads()
	if err != nil {
		h.cfg.Logger.Warn("Error getting active downloads from Transmission", "error", err)
		return
	}

	var downloadNodes []models.FileNode
	for _, name := range activeDownloads {
		downloadNodes = append(downloadNodes, models.FileNode{
			Name:          name,
			FormatSize:    "downloading",
			Icon:          "pi-cloud-download",
			Color:         "#2196F3",
			IsFolder:      false,
			IsDownloading: true,
		})
	}

	if err := h.templates.ExecuteTemplate(w, "rows.html", downloadNodes); err != nil {
		h.cfg.Logger.Error("Error executing rows template for active downloads", "error", err)
	}
}

func (h *Handler) HandleFolderContent(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	validatedPath, err := h.fileService.ValidatePath(path)
	if err != nil {
		h.cfg.Logger.Warn("Access denied or invalid path", "path", path, "error", err)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	levelStr := r.URL.Query().Get("level")
	level := 0
	if levelStr != "" {
		fmt.Sscanf(levelStr, "%d", &level)
	}

	nodes, err := h.fileService.ScanDirectory(validatedPath, level+1)
	if err != nil {
		h.cfg.Logger.Error("Error scanning folder content", "path", validatedPath, "error", err)
		http.Error(w, "Error scanning directory", http.StatusInternalServerError)
		return
	}

	for i := range nodes {
		h.fileService.FillFormatSize(&nodes[i])
	}

	if err := h.templates.ExecuteTemplate(w, "rows.html", nodes); err != nil {
		h.cfg.Logger.Error("Error executing rows template", "error", err)
	}
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	validatedPath, err := h.fileService.ValidatePath(path)
	if err != nil {
		h.cfg.Logger.Warn("Access denied or invalid path for download", "path", path, "error", err)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	stat, err := os.Stat(validatedPath)
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

		err := filepath.WalkDir(validatedPath, func(fpath string, d os.DirEntry, err error) error {
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

			relPath, err := filepath.Rel(validatedPath, fpath)
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
			h.cfg.Logger.Error("Error during zip download", "path", validatedPath, "error", err)
		}
		return
	} else {
		file, err := os.Open(validatedPath)
		if err != nil {
			h.cfg.Logger.Error("Error opening file for download", "path", validatedPath, "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
		w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	}
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if err := h.fileService.DeletePath(path); err != nil {
		h.cfg.Logger.Error("Error deleting path", "path", path, "error", err)
		http.Error(w, "The file could not be deleted (Read-only system?)", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func LoggingMiddleware(logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request", "method", r.Method, "url", r.URL.String())
		next(w, r)
	}
}
