// All rights reserved Samyak-Setu

package services

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// LocalStorageService implements StorageService using local filesystem storage.
type LocalStorageService struct {
	basePath string
}

// NewLocalStorageService creates a new LocalStorageService.
// It ensures the base upload directory exists.
func NewLocalStorageService(basePath string) (*LocalStorageService, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory %s: %w", basePath, err)
	}

	log.Printf("INFO: Local storage initialized at: %s", basePath)
	return &LocalStorageService{basePath: basePath}, nil
}

// SaveFile stores an uploaded file to the local filesystem.
// Returns the relative path to the saved file.
func (s *LocalStorageService) SaveFile(file *multipart.FileHeader, subDir string) (string, error) {
	// Create subdirectory if needed
	targetDir := filepath.Join(s.basePath, subDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Generate unique filename using timestamp
	ext := filepath.Ext(file.Filename)
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(targetDir, uniqueName)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("failed to write file: %w", writeErr)
			}
		}
		if readErr != nil {
			break
		}
	}

	// Return subDir/filename path
	return filepath.Join(subDir, uniqueName), nil
}

// SaveBytes creates a file from raw bytes on the local filesystem.
func (s *LocalStorageService) SaveBytes(data []byte, contentType, ext, subDir string) (string, error) {
	targetDir := filepath.Join(s.basePath, subDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(targetDir, uniqueName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write bytes to file: %w", err)
	}

	return filepath.Join(subDir, uniqueName), nil
}
