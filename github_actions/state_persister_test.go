package github_actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStatePersister(t *testing.T) {
	sp := NewStatePersister("test-session", "job-1", "./test-cache")
	
	if sp.sessionID != "test-session" {
		t.Errorf("Expected sessionID 'test-session', got '%s'", sp.sessionID)
	}
	
	if sp.matrixJobID != "job-1" {
		t.Errorf("Expected matrixJobID 'job-1', got '%s'", sp.matrixJobID)
	}
	
	if sp.cacheBaseDir != "./test-cache" {
		t.Errorf("Expected cacheBaseDir './test-cache', got '%s'", sp.cacheBaseDir)
	}
}

func TestGetCacheKey(t *testing.T) {
	sp := NewStatePersister("session-123", "job-5", "./cache")
	
	expected := "state-session-123-job-5"
	actual := sp.GetCacheKey()
	
	if actual != expected {
		t.Errorf("Expected cache key '%s', got '%s'", expected, actual)
	}
}

func TestGetSharedConfigKey(t *testing.T) {
	expected := "shared-config-latest"
	actual := GetSharedConfigKey()
	
	if actual != expected {
		t.Errorf("Expected shared config key '%s', got '%s'", expected, actual)
	}
}

func TestSaveAndRestoreState(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	recordingsDir := filepath.Join(tempDir, "recordings")
	restoreConfigDir := filepath.Join(tempDir, "restore-config")
	restoreRecordingsDir := filepath.Join(tempDir, "restore-recordings")
	
	// Create test files
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings directory: %v", err)
	}
	
	// Write test config file
	configContent := []byte("test config content")
	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Write test recording file
	recordingContent := []byte("test recording content")
	recordingFile := filepath.Join(recordingsDir, "recording.mp4")
	if err := os.WriteFile(recordingFile, recordingContent, 0644); err != nil {
		t.Fatalf("Failed to write recording file: %v", err)
	}
	
	// Create StatePersister
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// Save state
	if err := sp.SaveState(ctx, configDir, recordingsDir); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}
	
	// Verify manifest was created
	manifestPath := filepath.Join(cacheDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("Manifest file was not created")
	}
	
	// Restore state to different directories
	if err := sp.RestoreState(ctx, restoreConfigDir, restoreRecordingsDir); err != nil {
		t.Fatalf("RestoreState failed: %v", err)
	}
	
	// Verify restored config file
	restoredConfigFile := filepath.Join(restoreConfigDir, "config.json")
	restoredConfigContent, err := os.ReadFile(restoredConfigFile)
	if err != nil {
		t.Fatalf("Failed to read restored config file: %v", err)
	}
	
	if string(restoredConfigContent) != string(configContent) {
		t.Errorf("Restored config content mismatch: expected '%s', got '%s'",
			string(configContent), string(restoredConfigContent))
	}
	
	// Verify restored recording file
	restoredRecordingFile := filepath.Join(restoreRecordingsDir, "recording.mp4")
	restoredRecordingContent, err := os.ReadFile(restoredRecordingFile)
	if err != nil {
		t.Fatalf("Failed to read restored recording file: %v", err)
	}
	
	if string(restoredRecordingContent) != string(recordingContent) {
		t.Errorf("Restored recording content mismatch: expected '%s', got '%s'",
			string(recordingContent), string(restoredRecordingContent))
	}
}

func TestVerifyIntegrity(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Write test file
	content := []byte("test content for checksum")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Calculate checksum
	checksum, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}
	
	// Create manifest with correct checksum
	manifest := StateManifest{
		Files: []FileEntry{
			{
				Path:      "test.txt",
				Checksum:  checksum,
				Size:      int64(len(content)),
				Timestamp: time.Now(),
			},
		},
	}
	
	sp := NewStatePersister("test-session", "job-1", tempDir)
	
	// Verify integrity should pass
	if err := sp.VerifyIntegrity(manifest); err != nil {
		t.Errorf("VerifyIntegrity failed with correct checksum: %v", err)
	}
	
	// Create manifest with incorrect checksum
	badManifest := StateManifest{
		Files: []FileEntry{
			{
				Path:      "test.txt",
				Checksum:  "incorrect-checksum",
				Size:      int64(len(content)),
				Timestamp: time.Now(),
			},
		},
	}
	
	// Verify integrity should fail
	if err := sp.VerifyIntegrity(badManifest); err == nil {
		t.Error("VerifyIntegrity should have failed with incorrect checksum")
	}
}

func TestVerifyIntegrity_MissingFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create manifest for a file that doesn't exist
	manifest := StateManifest{
		Files: []FileEntry{
			{
				Path:      "missing.txt",
				Checksum:  "abc123",
				Size:      100,
				Timestamp: time.Now(),
			},
		},
	}
	
	sp := NewStatePersister("test-session", "job-1", tempDir)
	
	// Verify integrity should fail with cache miss
	err := sp.VerifyIntegrity(manifest)
	if err == nil {
		t.Error("VerifyIntegrity should have failed for missing file")
	}
	
	// Error should mention cache miss
	if err != nil && !contains(err.Error(), "missing from cache") {
		t.Errorf("Error should mention cache miss, got: %v", err)
	}
}

func TestVerifyIntegrity_SizeMismatch(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Write test file
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Calculate checksum
	checksum, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}
	
	// Create manifest with incorrect size
	manifest := StateManifest{
		Files: []FileEntry{
			{
				Path:      "test.txt",
				Checksum:  checksum,
				Size:      999, // Wrong size
				Timestamp: time.Now(),
			},
		},
	}
	
	sp := NewStatePersister("test-session", "job-1", tempDir)
	
	// Verify integrity should fail with size mismatch
	err = sp.VerifyIntegrity(manifest)
	if err == nil {
		t.Error("VerifyIntegrity should have failed for size mismatch")
	}
	
	// Error should mention size mismatch
	if err != nil && !contains(err.Error(), "size mismatch") {
		t.Errorf("Error should mention size mismatch, got: %v", err)
	}
}

func TestVerifyIntegrity_ChecksumMismatch(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Write test file
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Create manifest with incorrect checksum but correct size
	manifest := StateManifest{
		Files: []FileEntry{
			{
				Path:      "test.txt",
				Checksum:  "0000000000000000000000000000000000000000000000000000000000000000",
				Size:      int64(len(content)),
				Timestamp: time.Now(),
			},
		},
	}
	
	sp := NewStatePersister("test-session", "job-1", tempDir)
	
	// Verify integrity should fail with checksum mismatch
	err := sp.VerifyIntegrity(manifest)
	if err == nil {
		t.Error("VerifyIntegrity should have failed for checksum mismatch")
	}
	
	// Error should mention checksum mismatch
	if err != nil && !contains(err.Error(), "checksum mismatch") {
		t.Errorf("Error should mention checksum mismatch, got: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr)+1 && s[1:len(substr)+1] == substr ||
		// Simple contains check
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}

func TestRestoreStateWithMissingCache(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// Attempt to restore state when cache doesn't exist
	err := sp.RestoreState(ctx, configDir, "")
	if err == nil {
		t.Error("RestoreState should return error when cache is missing")
	}
	
	// Error should be a cache miss error
	if !IsCacheMiss(err) {
		t.Errorf("Error should be a cache miss error, got: %v", err)
	}
	
	// Error message should indicate cache miss
	if err != nil {
		t.Logf("Expected error received: %v", err)
	}
}

func TestIsCacheMiss(t *testing.T) {
	// Test with actual cache miss error
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	err := sp.RestoreState(ctx, configDir, "")
	if err == nil {
		t.Fatal("Expected error from RestoreState with missing cache")
	}
	
	if !IsCacheMiss(err) {
		t.Errorf("IsCacheMiss should return true for cache miss error, got false")
	}
	
	// Test with non-cache-miss error
	otherErr := fmt.Errorf("some other error")
	if IsCacheMiss(otherErr) {
		t.Errorf("IsCacheMiss should return false for non-cache-miss error")
	}
	
	// Test with nil error
	if IsCacheMiss(nil) {
		t.Errorf("IsCacheMiss should return false for nil error")
	}
}

func TestSaveStateWithNonexistentDirectories(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// Save state with nonexistent directories (should not fail)
	err := sp.SaveState(ctx, filepath.Join(tempDir, "nonexistent-config"), filepath.Join(tempDir, "nonexistent-recordings"))
	if err != nil {
		t.Errorf("SaveState should not fail with nonexistent directories: %v", err)
	}
	
	// Manifest should still be created
	manifestPath := filepath.Join(cacheDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Manifest file should be created even with nonexistent source directories")
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Calculate checksum
	checksum1, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}
	
	// Calculate checksum again - should be identical
	checksum2, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}
	
	if checksum1 != checksum2 {
		t.Errorf("Checksums should be identical: %s != %s", checksum1, checksum2)
	}
	
	// Checksum should be 64 characters (SHA-256 hex)
	if len(checksum1) != 64 {
		t.Errorf("Checksum should be 64 characters, got %d", len(checksum1))
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")
	
	content := []byte("test content for copy")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	
	// Copy file
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}
	
	// Verify destination file exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Fatal("Destination file was not created")
	}
	
	// Verify content matches
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	
	if string(dstContent) != string(content) {
		t.Errorf("Content mismatch: expected '%s', got '%s'", string(content), string(dstContent))
	}
}

func TestIncrementalCacheUpdates(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	
	// Create test files
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Write initial config file
	configFile := filepath.Join(configDir, "config.json")
	initialContent := []byte("initial config content")
	if err := os.WriteFile(configFile, initialContent, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create StatePersister
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// First save - should save all files
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("First SaveState failed: %v", err)
	}
	
	// Verify checksum cache is populated
	if len(sp.fileChecksumCache) != 1 {
		t.Errorf("Expected 1 entry in checksum cache, got %d", len(sp.fileChecksumCache))
	}
	
	// Second save without changes - should skip unchanged files
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("Second SaveState failed: %v", err)
	}
	
	// Modify the config file
	modifiedContent := []byte("modified config content")
	if err := os.WriteFile(configFile, modifiedContent, 0644); err != nil {
		t.Fatalf("Failed to modify config file: %v", err)
	}
	
	// Third save with changes - should update changed file
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("Third SaveState failed: %v", err)
	}
	
	// Verify the cached file has the new content
	cachedFile := filepath.Join(cacheDir, "config", "config.json")
	cachedContent, err := os.ReadFile(cachedFile)
	if err != nil {
		t.Fatalf("Failed to read cached file: %v", err)
	}
	
	if string(cachedContent) != string(modifiedContent) {
		t.Errorf("Cached content mismatch: expected '%s', got '%s'",
			string(modifiedContent), string(cachedContent))
	}
}

func TestIncrementalCacheUpdatesMultipleFiles(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	
	// Create test files
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Write multiple config files
	file1 := filepath.Join(configDir, "config1.json")
	file2 := filepath.Join(configDir, "config2.json")
	file3 := filepath.Join(configDir, "config3.json")
	
	content1 := []byte("config 1 content")
	content2 := []byte("config 2 content")
	content3 := []byte("config 3 content")
	
	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}
	if err := os.WriteFile(file3, content3, 0644); err != nil {
		t.Fatalf("Failed to write file3: %v", err)
	}
	
	// Create StatePersister
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// First save - should save all files
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("First SaveState failed: %v", err)
	}
	
	// Verify checksum cache has 3 entries
	if len(sp.fileChecksumCache) != 3 {
		t.Errorf("Expected 3 entries in checksum cache, got %d", len(sp.fileChecksumCache))
	}
	
	// Modify only file2
	modifiedContent2 := []byte("modified config 2 content")
	if err := os.WriteFile(file2, modifiedContent2, 0644); err != nil {
		t.Fatalf("Failed to modify file2: %v", err)
	}
	
	// Second save - should only update file2
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("Second SaveState failed: %v", err)
	}
	
	// Verify file1 and file3 are unchanged in cache
	cachedFile1 := filepath.Join(cacheDir, "config", "config1.json")
	cachedContent1, err := os.ReadFile(cachedFile1)
	if err != nil {
		t.Fatalf("Failed to read cached file1: %v", err)
	}
	if string(cachedContent1) != string(content1) {
		t.Errorf("File1 should be unchanged")
	}
	
	// Verify file2 is updated in cache
	cachedFile2 := filepath.Join(cacheDir, "config", "config2.json")
	cachedContent2, err := os.ReadFile(cachedFile2)
	if err != nil {
		t.Fatalf("Failed to read cached file2: %v", err)
	}
	if string(cachedContent2) != string(modifiedContent2) {
		t.Errorf("File2 should be updated: expected '%s', got '%s'",
			string(modifiedContent2), string(cachedContent2))
	}
}

func TestIncrementalCacheUpdatesAfterRestore(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	restoreConfigDir := filepath.Join(tempDir, "restore-config")
	
	// Create test files
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Write initial config file
	configFile := filepath.Join(configDir, "config.json")
	initialContent := []byte("initial config content")
	if err := os.WriteFile(configFile, initialContent, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create first StatePersister and save state
	sp1 := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	if err := sp1.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}
	
	// Create second StatePersister (simulating new workflow run)
	sp2 := NewStatePersister("test-session", "job-1", cacheDir)
	
	// Restore state - should populate checksum cache
	if err := sp2.RestoreState(ctx, restoreConfigDir, ""); err != nil {
		t.Fatalf("RestoreState failed: %v", err)
	}
	
	// Verify checksum cache is populated
	if len(sp2.fileChecksumCache) != 1 {
		t.Errorf("Expected 1 entry in checksum cache after restore, got %d", len(sp2.fileChecksumCache))
	}
	
	// Modify the restored file
	restoredFile := filepath.Join(restoreConfigDir, "config.json")
	modifiedContent := []byte("modified config content")
	if err := os.WriteFile(restoredFile, modifiedContent, 0644); err != nil {
		t.Fatalf("Failed to modify restored file: %v", err)
	}
	
	// Save state again - should detect the change
	if err := sp2.SaveState(ctx, restoreConfigDir, ""); err != nil {
		t.Fatalf("SaveState after restore failed: %v", err)
	}
	
	// Verify the cached file has the new content
	cachedFile := filepath.Join(cacheDir, "config", "config.json")
	cachedContent, err := os.ReadFile(cachedFile)
	if err != nil {
		t.Fatalf("Failed to read cached file: %v", err)
	}
	
	if string(cachedContent) != string(modifiedContent) {
		t.Errorf("Cached content should be updated: expected '%s', got '%s'",
			string(modifiedContent), string(cachedContent))
	}
}

func TestIncrementalCacheUpdatesNewFiles(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	
	// Create test files
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Write initial config file
	file1 := filepath.Join(configDir, "config1.json")
	content1 := []byte("config 1 content")
	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	
	// Create StatePersister
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// First save - should save file1
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("First SaveState failed: %v", err)
	}
	
	// Verify checksum cache has 1 entry
	if len(sp.fileChecksumCache) != 1 {
		t.Errorf("Expected 1 entry in checksum cache, got %d", len(sp.fileChecksumCache))
	}
	
	// Add a new file
	file2 := filepath.Join(configDir, "config2.json")
	content2 := []byte("config 2 content")
	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}
	
	// Second save - should add the new file
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("Second SaveState failed: %v", err)
	}
	
	// Verify checksum cache has 2 entries
	if len(sp.fileChecksumCache) != 2 {
		t.Errorf("Expected 2 entries in checksum cache, got %d", len(sp.fileChecksumCache))
	}
	
	// Verify both files are in cache
	cachedFile1 := filepath.Join(cacheDir, "config", "config1.json")
	cachedFile2 := filepath.Join(cacheDir, "config", "config2.json")
	
	if _, err := os.Stat(cachedFile1); os.IsNotExist(err) {
		t.Error("File1 should exist in cache")
	}
	
	if _, err := os.Stat(cachedFile2); os.IsNotExist(err) {
		t.Error("File2 should exist in cache")
	}
}

func TestIncrementalCacheUpdatesEmptyDirectory(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	configDir := filepath.Join(tempDir, "config")
	
	// Create empty config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create StatePersister
	sp := NewStatePersister("test-session", "job-1", cacheDir)
	ctx := context.Background()
	
	// Save state with empty directory - should not fail
	if err := sp.SaveState(ctx, configDir, ""); err != nil {
		t.Fatalf("SaveState with empty directory failed: %v", err)
	}
	
	// Verify checksum cache is empty
	if len(sp.fileChecksumCache) != 0 {
		t.Errorf("Expected 0 entries in checksum cache for empty directory, got %d", len(sp.fileChecksumCache))
	}
	
	// Verify manifest was still created
	manifestPath := filepath.Join(cacheDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Manifest file should be created even with empty directory")
	}
}
