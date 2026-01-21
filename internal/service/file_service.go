package service

import (
	"ddownload/internal/models"
	"ddownload/internal/utils"
	"os"
	"path/filepath"
)

func ScanDirectory(rootPath string, level int) ([]models.FileNode, error) {
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
			node.TotalSize = CalculateDirSize(fullPath)
		} else {
			node.TotalSize = info.Size()
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func CalculateDirSize(path string) int64 {
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
	return size
}

func FillFormatSize(node *models.FileNode) {
	node.FormatSize = utils.FormatBytes(node.TotalSize)

	for i := range node.Children {
		FillFormatSize(&node.Children[i])
	}
}
