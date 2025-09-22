package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"incident-management-system/internal/database"
	"incident-management-system/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProcessingService is a mock implementation of the processing service
type MockProcessingService struct {
	ProcessUploadFunc       func(ctx context.Context, uploadID string) (interface{}, error)
	GetProcessingStatusFunc func(ctx context.Context, uploadID string) (interface{}, error)
}

func (m *MockProcessingService) ProcessUpload(ctx context.Context, uploadID string) (interface{}, error) {
	if m.ProcessUploadFunc != nil {
		return m.ProcessUploadFunc(ctx, uploadID)
	}
	return nil, nil
}

func (m *MockProcessingService) GetProcessingStatus(ctx context.Context, uploadID string) (interface{}, error) {
	if m.GetProcessingStatusFunc != nil {
		return m.GetProcessingStatusFunc(ctx, uploadID)
	}
	return nil, nil
}

// createTestDB creates a test database connection
func createTestDB(t *testing.T) *sql.DB {
	config := &database.Config{
		DatabasePath: ":memory:",
	}

	dbWrapper, err := database.NewDB(config)
	require.NoError(t, err, "Failed to create test database")

	err = dbWrapper.InitializeDatabase()
	require.NoError(t, err, "Failed to initialize test database")

	t.Cleanup(func() {
		dbWrapper.Close()
	})

	return dbWrapper.GetConnection()
}

// createTestFile creates a test file for upload testing
func createTestFile(t *testing.T, content string) (*os.File, string) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.xlsx")

	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file")

	file, err := os.Open(filePath)
	require.NoError(t, err, "Failed to open test file")

	t.Cleanup(func() {
		file.Close()
	})

	return file, filepath.Base(filePath)
}

// createMultipartForm creates a multipart form with a file
func createMultipartForm(t *testing.T, filename, content string) (*bytes.Buffer, *multipart.Writer) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err, "Failed to create form file")

	_, err = io.WriteString(part, content)
	require.NoError(t, err, "Failed to write file content")

	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	return body, writer
}

func TestUploadHandler_UploadFile(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDB(t)

	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	mockService := new(MockProcessingService)
	handler := NewUploadHandler(db, fileStore, mockService)

	// Test case: successful upload
	t.Run("successful upload", func(t *testing.T) {
		body, writer := createMultipartForm(t, "test.xlsx", "test content")
		req := httptest.NewRequest("POST", "/uploads", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Create response recorder
		w := httptest.NewRecorder()

		// Create gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.UploadFile(c)

		// Check response
		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "File uploaded successfully", response["message"])
	})

	// Test case: no file provided
	t.Run("no file provided", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/uploads", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Create gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.UploadFile(c)

		// Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error structure
		assert.Equal(t, "MISSING_FILE", response["code"])
		assert.Contains(t, response["user_message"], "Please select a file to upload")
	})
}

func TestUploadHandler_UploadFile_LargeFile(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDB(t)

	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	mockService := new(MockProcessingService)
	handler := NewUploadHandler(db, fileStore, mockService)

	// Create a large file content (51MB)
	largeContent := strings.Repeat("a", 51<<20) // 51MB
	body, writer := createMultipartForm(t, "large.xlsx", largeContent)

	// Create request
	req := httptest.NewRequest("POST", "/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create response recorder
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.UploadFile(c)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check error structure
	assert.Equal(t, "FILE_TOO_LARGE", response["code"])
	assert.Contains(t, response["user_message"], "file is too large")
}

func TestUploadHandler_GetUploads(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDB(t)

	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	mockService := new(MockProcessingService)
	handler := NewUploadHandler(db, fileStore, mockService)

	// First, create a test upload using the handler
	body, writer := createMultipartForm(t, "test.xlsx", "test content")
	req := httptest.NewRequest("POST", "/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.UploadFile(c)

	// Verify upload was successful
	assert.Equal(t, http.StatusCreated, w.Code)

	// Create request to get uploads
	req = httptest.NewRequest("GET", "/uploads", nil)
	w = httptest.NewRecorder()

	// Create gin context
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetUploads(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	uploadsData, ok := response["uploads"].([]interface{})
	assert.True(t, ok, "Uploads should be an array")
	assert.Greater(t, len(uploadsData), 0, "Should return at least one upload")
}

func TestUploadHandler_GetUpload(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDB(t)

	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	mockService := new(MockProcessingService)
	handler := NewUploadHandler(db, fileStore, mockService)

	// First, create a test upload using the handler
	body, writer := createMultipartForm(t, "test.xlsx", "test content")
	req := httptest.NewRequest("POST", "/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.UploadFile(c)

	// Verify upload was successful
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse the response to get the upload ID
	var uploadResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	require.NoError(t, err)

	uploadData, ok := uploadResponse["upload"].(map[string]interface{})
	require.True(t, ok, "Upload should be an object")

	uploadID, ok := uploadData["id"].(string)
	require.True(t, ok, "Upload ID should be a string")

	// Test cases
	tests := []struct {
		name         string
		uploadID     string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful get upload",
			uploadID:    uploadID,
			expectError: false,
		},
		{
			name:         "upload not found",
			uploadID:     "non-existent-id",
			expectError:  true,
			errorMessage: "Upload not found",
		},
		{
			name:         "missing upload ID",
			uploadID:     "",
			expectError:  true,
			errorMessage: "Upload ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := "/uploads"
			if tt.uploadID != "" {
				url = fmt.Sprintf("/uploads/%s", tt.uploadID)
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			if tt.uploadID != "" {
				c.Params = []gin.Param{{Key: "id", Value: tt.uploadID}}
			}

			// Execute handler
			handler.GetUpload(c)

			// Check response
			if tt.expectError {
				if tt.errorMessage == "Upload not found" {
					assert.Equal(t, http.StatusNotFound, w.Code)
				} else {
					assert.Equal(t, http.StatusBadRequest, w.Code)
				}
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], tt.errorMessage)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				uploadData, ok := response["upload"].(map[string]interface{})
				assert.True(t, ok, "Upload should be an object")
				assert.Equal(t, uploadID, uploadData["id"])
			}
		})
	}
}

func TestUploadHandler_ProcessUpload(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDB(t)

	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	mockService := new(MockProcessingService)
	handler := NewUploadHandler(db, fileStore, mockService)

	// First, create a test upload using the handler
	body, writer := createMultipartForm(t, "test.xlsx", "test content")
	req := httptest.NewRequest("POST", "/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.UploadFile(c)

	// Verify upload was successful
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse the response to get the upload ID
	var uploadResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	require.NoError(t, err)

	uploadData, ok := uploadResponse["upload"].(map[string]interface{})
	require.True(t, ok, "Upload should be an object")

	uploadID, ok := uploadData["id"].(string)
	require.True(t, ok, "Upload ID should be a string")

	// Test cases
	tests := []struct {
		name         string
		uploadID     string
		setupMock    func()
		expectError  bool
		errorMessage string
	}{
		{
			name:     "successful process upload",
			uploadID: uploadID,
			setupMock: func() {
				mockService.ProcessUploadFunc = func(ctx context.Context, uploadID string) (interface{}, error) {
					return nil, nil
				}
			},
			expectError: false,
		},
		{
			name:         "upload not found",
			uploadID:     "non-existent-id",
			setupMock:    func() {},
			expectError:  true,
			errorMessage: "Upload not found",
		},
		{
			name:         "missing upload ID",
			uploadID:     "",
			setupMock:    func() {},
			expectError:  true,
			errorMessage: "Upload ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.setupMock()

			// Create request
			url := "/uploads/process"
			if tt.uploadID != "" {
				url = fmt.Sprintf("/uploads/%s/process", tt.uploadID)
			}
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			if tt.uploadID != "" {
				c.Params = []gin.Param{{Key: "id", Value: tt.uploadID}}
			}

			// Execute handler
			handler.ProcessUpload(c)

			// Check response
			if tt.expectError {
				if tt.errorMessage == "Upload not found" {
					assert.Equal(t, http.StatusNotFound, w.Code)
				} else {
					assert.Equal(t, http.StatusBadRequest, w.Code)
				}
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], tt.errorMessage)
			} else {
				assert.Equal(t, http.StatusAccepted, w.Code)
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "Processing started", response["message"])
			}
		})
	}
}
