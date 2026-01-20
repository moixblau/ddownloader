package utils

import (
	"fmt"
	"path/filepath"
)

func GetFileMeta(path string, isDir bool) (string, string) {
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

func FormatBytes(b int64) string {
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
