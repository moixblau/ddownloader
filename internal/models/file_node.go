package models

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
