package service

import (
	"ddownload/internal/models"
	"ddownload/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileService struct {
	DataDir   string
	sizeCache sync.Map
}

type cacheItem struct {
	size      int64
	timestamp time.Time
}

func NewFileService(dataDir string) *FileService {
	return &FileService{
		DataDir: dataDir,
	}
}

func (s *FileService) ScanDirectory(rootPath string, level int) ([]models.FileNode, error) {
	var nodes []models.FileNode

	files, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file.Name())
		info, _ := file.Info()

		icon, color := utils.GetFileMeta(file.Name(), file.IsDir())

		node := models.FileNode{
			Name:     file.Name(),
			Path:     fullPath,
			IsFolder: file.IsDir(),
			Icon:     icon,
			Color:    color,
			Level:    level,
		}

		if file.IsDir() {
			node.TotalSize = s.CalculateDirSize(fullPath)
		} else {
			if info != nil {
				node.TotalSize = info.Size()
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (s *FileService) SearchFiles(rootPath, query string) ([]models.FileNode, error) {
	var nodes []models.FileNode
	query = strings.ToLower(query)

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Ignorar errores en archivos individuales
		}

		if path == rootPath {
			return nil
		}

		if strings.Contains(strings.ToLower(d.Name()), query) {
			info, _ := d.Info()
			icon, color := utils.GetFileMeta(d.Name(), d.IsDir())

			rel, _ := filepath.Rel(rootPath, path)
			level := strings.Count(rel, string(os.PathSeparator))

			node := models.FileNode{
				Name:     d.Name(),
				Path:     path,
				IsFolder: d.IsDir(),
				Icon:     icon,
				Color:    color,
				Level:    level,
			}

			if d.IsDir() {
				node.TotalSize = s.CalculateDirSize(path)
			} else if info != nil {
				node.TotalSize = info.Size()
			}

			nodes = append(nodes, node)
		}
		return nil
	})

	return nodes, err
}

func (s *FileService) CalculateDirSize(path string) int64 {
	if val, ok := s.sizeCache.Load(path); ok {
		item := val.(cacheItem)
		if time.Since(item.timestamp) < 5*time.Minute {
			return item.size
		}
	}

	var size int64
	filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})

	s.sizeCache.Store(path, cacheItem{size: size, timestamp: time.Now()})
	return size
}

func (s *FileService) FillFormatSize(node *models.FileNode) {
	node.FormatSize = utils.FormatBytes(node.TotalSize)

	for i := range node.Children {
		s.FillFormatSize(&node.Children[i])
	}
}

func (s *FileService) ValidatePath(path string) (string, error) {
	absDataDir, err := filepath.Abs(s.DataDir)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absDataDir, absPath)
	if err != nil || (len(rel) >= 2 && rel[:2] == "..") {
		return "", os.ErrPermission
	}

	return absPath, nil
}

func (s *FileService) DeletePath(path string) error {
	validatedPath, err := s.ValidatePath(path)
	if err != nil {
		return err
	}
	return os.RemoveAll(validatedPath)
}
