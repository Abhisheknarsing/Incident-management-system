package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileStore handles file storage operations
type FileStore struct {
	uploadDir string
}

// NewFileStore creates a new FileStore instance
func NewFileStore(uploadDir string) *FileStore {
	return &FileStore{
		uploadDir: uploadDir,
	}
}

// SaveUploadedFile saves an uploaded file with a unique name
func (fs *FileStore) SaveUploadedFile(file *multipart.FileHeader) (string, string, error) {
	// Validate file extension
	if !fs.isValidExcelFile(file.Filename) {
		return "", "", fmt.Errorf("invalid file format: only .xlsx and .xls files are supported")
	}

	// Generate unique filename
	uniqueFilename := fs.generateUniqueFilename(file.Filename)
	filePath := filepath.Join(fs.uploadDir, uniqueFilename)

	// Ensure upload directory exists
	if err := os.MkdirAll(fs.uploadDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		// Clean up on failure
		os.Remove(filePath)
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	return uniqueFilename, filePath, nil
}

// DeleteFile removes a file from storage
func (fs *FileStore) DeleteFile(filename string) error {
	filePath := filepath.Join(fs.uploadDir, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", filename, err)
	}
	return nil
}

// GetFilePath returns the full path to a stored file
func (fs *FileStore) GetFilePath(filename string) string {
	return filepath.Join(fs.uploadDir, filename)
}

// isValidExcelFile checks if the file has a valid Excel extension
func (fs *FileStore) isValidExcelFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".xlsx" || ext == ".xls"
}

// generateUniqueFilename creates a unique filename while preserving the extension
func (fs *FileStore) generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Format("20060102_150405")
	uuid := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s%s", timestamp, uuid, ext)
}