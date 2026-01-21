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

	svc := NewFileService(tmpDir)
	got := svc.CalculateDirSize(tmpDir)
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

	svc := NewFileService("/tmp")
	svc.FillFormatSize(node)

	if node.FormatSize != "1.00 KB" {
		t.Errorf("node.FormatSize = %q; want %q", node.FormatSize, "1.00 KB")
	}
	if node.Children[0].FormatSize != "512 B" {
		t.Errorf("node.Children[0].FormatSize = %q; want %q", node.Children[0].FormatSize, "512 B")
	}
}

func TestValidatePath(t *testing.T) {
	dataDir, _ := filepath.Abs("test_data")
	svc := NewFileService(dataDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"Valid path", filepath.Join(dataDir, "file.txt"), false},
		{"Path outside", "/etc/passwd", true},
		{"Relative outside", filepath.Join(dataDir, "..", "passwd"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testsearch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create structure:
	// root/file_at_root.txt
	// root/folder1/file_in_folder.txt
	// root/folder2/subfolder/deep_file.log

	os.WriteFile(filepath.Join(tmpDir, "file_at_root.txt"), []byte("a"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "folder1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "folder1", "file_in_folder.txt"), []byte("b"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "folder2", "subfolder"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "folder2", "subfolder", "deep_file.log"), []byte("c"), 0644)

	svc := NewFileService(tmpDir)

	tests := []struct {
		query     string
		wantCount int
	}{
		{"file_at_root", 1},
		{"folder1", 1},
		{"subfolder", 1},
		{"file_in_folder", 1},
		{"deep", 1},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			nodes, err := svc.SearchFiles(tmpDir, tt.query)
			if err != nil {
				t.Fatalf("SearchFiles failed: %v", err)
			}
			if len(nodes) != tt.wantCount {
				t.Errorf("SearchFiles(%q) got %d nodes, want %d", tt.query, len(nodes), tt.wantCount)
			}
		})
	}
}

func TestStrangeCharacters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "teststrange")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	strangeNames := []string{
		"archivo con espacios.txt",
		"niño.txt",
		"canción.mp3",
		"🔥emoji.png",
		"special!@#$%^&()_+-=.txt",
		"日本語.txt",
	}

	for _, name := range strangeNames {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create file %q: %v", name, err)
		}
	}

	svc := NewFileService(tmpDir)
	nodes, err := svc.ScanDirectory(tmpDir, 0)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(nodes) != len(strangeNames) {
		t.Errorf("got %d nodes, want %d", len(nodes), len(strangeNames))
	}

	// Verify each name is present
	foundNames := make(map[string]bool)
	for _, node := range nodes {
		foundNames[node.Name] = true
	}

	for _, name := range strangeNames {
		if !foundNames[name] {
			t.Errorf("name %q not found in ScanDirectory results", name)
		}
	}

	// Test SearchFiles with strange characters
	t.Run("SearchStrangeCharacters", func(t *testing.T) {
		query := "niño"
		nodes, err := svc.SearchFiles(tmpDir, query)
		if err != nil {
			t.Fatalf("SearchFiles failed: %v", err)
		}
		if len(nodes) != 1 || nodes[0].Name != "niño.txt" {
			t.Errorf("SearchFiles(%q) failed to find the file", query)
		}

		query = "🔥"
		nodes, err = svc.SearchFiles(tmpDir, query)
		if err != nil {
			t.Fatalf("SearchFiles failed: %v", err)
		}
		if len(nodes) != 1 || nodes[0].Name != "🔥emoji.png" {
			t.Errorf("SearchFiles(%q) failed to find the file", query)
		}
	})
}

func TestManyFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testmanyfiles")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	count := 1000
	for i := 0; i < count; i++ {
		name := filepath.Join(tmpDir, "file_"+string(rune(i))+".txt")
		// Use a more predictable naming to avoid filesystem issues with some control characters if rune(i) is weird
		name = filepath.Join(tmpDir, "file_"+os.ExpandEnv(string(rune(65+(i%26))))+"_"+string(rune(48+(i%10)))+"_"+string(rune(i))+".txt")
		// Actually simpler:
		name = filepath.Join(tmpDir, "file_"+string(rune(i%1000))+".txt")
		// Wait, some runes might be invalid for filenames or be path separators. Let's use simple numbers.
		name = filepath.Join(tmpDir, "file_"+string(rune(48+(i/100)))+string(rune(48+(i/10)%10))+string(rune(48+(i%10)))+".txt")

		err := os.WriteFile(name, []byte("a"), 0644)
		if err != nil {
			t.Fatalf("failed to create file %d: %v", i, err)
		}
	}

	svc := NewFileService(tmpDir)
	nodes, err := svc.ScanDirectory(tmpDir, 0)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(nodes) != count {
		t.Errorf("got %d nodes, want %d", len(nodes), count)
	}

	// Test CalculateDirSize with many files
	size := svc.CalculateDirSize(tmpDir)
	if size != int64(count) {
		t.Errorf("CalculateDirSize() = %d; want %d", size, count)
	}
}
