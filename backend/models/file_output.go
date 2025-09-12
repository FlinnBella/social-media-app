package models

import "path/filepath"

// FileOutput represents a temporary file with its metadata
type FileOutput struct {
	FilePath string // Absolute path to the file
	FileName string // Just the filename (without path)
	TmpDir   string // Temporary directory containing the file
}

// NewFileOutput creates a new FileOutput instance
func NewFileOutput(filePath, tmpDir string) *FileOutput {
	return &FileOutput{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
		TmpDir:   tmpDir,
	}
}
