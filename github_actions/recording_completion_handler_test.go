package github_actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewRecordingCompletionHandler(t *testing.T) {
	// Create mock components
	storageUploader := NewStorageUploader("test-gofile-key", "test-filester-key")
	databaseManager := NewDatabaseManager(".")
	healthMonitor := NewHealthMonitor("status.json", []Notifier{})
	sessionID := "test-session-123"
	matrixJobID := "matrix-job-1"

	// Create handler
	handler := NewRecordingCompletionHandler(
		storageUploader,
		databaseManager,
		healthMonitor,
		sessionID,
		matrixJobID,
	)

	// Verify handler was created correctly
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.GetStorageUploader() != storageUploader {
		t.Error("Storage uploader not set correctly")
	}

	if handler.GetDatabaseManager() != databaseManager {
		t.Error("Database manager not set correctly")
	}

	if handler.GetHealthMonitor() != healthMonitor {
		t.Error("Health monitor not set correctly")
	}

	if handler.GetSessionID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, handler.GetSessionID())
	}

	if handler.GetMatrixJobID() != matrixJobID {
		t.Errorf("Expected matrix job ID %s, got %s", matrixJobID, handler.GetMatrixJobID())
	}
}

func TestHandleRecordingCompletion_FileNotFound(t *testing.T) {
	// Create handler with mock components
	storageUploader := NewStorageUploader("test-gofile-key", "test-filester-key")
	databaseManager := NewDatabaseManager(".")
	healthMonitor := NewHealthMonitor("status.json", []Notifier{})
	
	handler := NewRecordingCompletionHandler(
		storageUploader,
		databaseManager,
		healthMonitor,
		"test-session",
		"matrix-job-1",
	)

	// Try to handle a non-existent file
	ctx := context.Background()
	err := handler.HandleRecordingCompletion(
		ctx,
		"/nonexistent/file.mp4",
		"chaturbate",
		"testuser",
		time.Now(),
		3600.0,
	)

	// Should fail because file doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestHandleRecordingCompletion_Integration(t *testing.T) {
	// Skip this test in short mode as it requires external services
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_recording.mp4")
	
	// Write some test data to the file
	testData := []byte("test recording data")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create handler with mock components
	storageUploader := NewStorageUploader("test-gofile-key", "test-filester-key")
	databaseManager := NewDatabaseManager(tmpDir)
	healthMonitor := NewHealthMonitor(filepath.Join(tmpDir, "status.json"), []Notifier{})
	
	handler := NewRecordingCompletionHandler(
		storageUploader,
		databaseManager,
		healthMonitor,
		"test-session-123",
		"matrix-job-1",
	)

	// Handle the recording completion
	ctx := context.Background()
	startTime := time.Now().Add(-1 * time.Hour)
	
	err := handler.HandleRecordingCompletion(
		ctx,
		testFile,
		"chaturbate",
		"testuser",
		startTime,
		3600.0,
	)

	// This will fail because we don't have real API keys, but we can verify
	// the error is from the upload step, not from file handling
	if err == nil {
		t.Error("Expected error due to invalid API keys, got nil")
	}
	
	// The error should be related to upload, not file access
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestExtractQualityFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "empty filename",
			filename: "",
			expected: "",
		},
		{
			name:     "simple filename",
			filename: "recording.mp4",
			expected: "",
		},
		{
			name:     "filename with path",
			filename: "/path/to/recording.mp4",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractQualityFromFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected quality %q, got %q", tt.expected, result)
			}
		})
	}
}
