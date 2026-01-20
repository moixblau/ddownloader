package utils

import (
	"testing"
)

func TestGetFileMeta(t *testing.T) {
	tests := []struct {
		path     string
		isDir    bool
		wantIcon string
		wantCol  string
	}{
		{"folder", true, "pi-folder", "#fbc02d"},
		{"video.mp4", false, "pi-video", "#ff5722"},
		{"image.png", false, "pi-image", "#4caf50"},
		{"doc.pdf", false, "pi-file-pdf", "#f44336"},
		{"archive.zip", false, "pi-file-archive", "#9c27b0"},
		{"unknown.txt", false, "pi-file", "var(--primary-color)"},
	}

	for _, tt := range tests {
		gotIcon, gotCol := GetFileMeta(tt.path, tt.isDir)
		if gotIcon != tt.wantIcon || gotCol != tt.wantCol {
			t.Errorf("GetFileMeta(%q, %v) = (%q, %q); want (%q, %q)",
				tt.path, tt.isDir, gotIcon, gotCol, tt.wantIcon, tt.wantCol)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
		{2048, "2.00 KB"},
		{1536, "1.50 KB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q; want %q", tt.bytes, got, tt.want)
		}
	}
}
