package service

import (
	"ddownload/internal/models"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateDirSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testsize")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "file1.txt")
	content1 := []byte("hello") // 5 bytes
	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(subDir, "file2.txt")
	content2 := []byte("world!!") // 7 bytes
	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatal(err)
	}

	got := CalculateDirSize(tmpDir)
	want := int64(12)

	if got != want {
		t.Errorf("CalculateDirSize() = %d; want %d", got, want)
	}
}

func TestFillFormatSize(t *testing.T) {
	node := &models.FileNode{
		TotalSize: 1024,
		Children: []models.FileNode{
			{TotalSize: 512},
		},
	}

	FillFormatSize(node)

	if node.FormatSize != "1.00 KB" {
		t.Errorf("node.FormatSize = %q; want %q", node.FormatSize, "1.00 KB")
	}
	if node.Children[0].FormatSize != "512 B" {
		t.Errorf("node.Children[0].FormatSize = %q; want %q", node.Children[0].FormatSize, "512 B")
	}
}
