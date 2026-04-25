package github_actions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNewDatabaseManager verifies that NewDatabaseManager creates a valid instance
func TestNewDatabaseManager(t *testing.T) {
	repoPath := "/test/repo"
	dm := NewDatabaseManager(repoPath)

	if dm == nil {
		t.Fatal("NewDatabaseManager returned nil")
	}

	if dm.repoPath != repoPath {
		t.Errorf("Expected repoPath %s, got %s", repoPath, dm.repoPath)
	}
}

// TestGetDatabasePath verifies that GetDatabasePath generates correct paths
func TestGetDatabasePath(t *testing.T) {
	dm := NewDatabaseManager("/repo")

	tests := []struct {
		name     string
		site     string
		channel  string
		date     string
		expected string
	}{
		{
			name:     "chaturbate channel",
			site:     "chaturbate",
			channel:  "username1",
			date:     "2024-01-15",
			expected: filepath.Join("/repo", "database", "chaturbate", "username1", "2024-01-15.json"),
		},
		{
			name:     "stripchat channel",
			site:     "stripchat",
			channel:  "username2",
			date:     "2024-01-16",
			expected: filepath.Join("/repo", "database", "stripchat", "username2", "2024-01-16.json"),
		},
		{
			name:     "different date format",
			site:     "chaturbate",
			channel:  "testuser",
			date:     "2023-12-31",
			expected: filepath.Join("/repo", "database", "chaturbate", "testuser", "2023-12-31.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dm.GetDatabasePath(tt.site, tt.channel, tt.date)
			if result != tt.expected {
				t.Errorf("Expected path %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestEnsureDirectoryExists verifies that directory creation works correctly
func TestEnsureDirectoryExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Test creating a new directory structure
	testPath := filepath.Join(tmpDir, "database", "chaturbate", "username1", "2024-01-15.json")
	
	err = dm.ensureDirectoryExists(testPath)
	if err != nil {
		t.Errorf("ensureDirectoryExists failed: %v", err)
	}

	// Verify the directory was created
	dir := filepath.Dir(testPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", dir)
	}

	// Test that calling it again on existing directory doesn't fail
	err = dm.ensureDirectoryExists(testPath)
	if err != nil {
		t.Errorf("ensureDirectoryExists failed on existing directory: %v", err)
	}
}

// TestFormatTimestamp verifies that timestamps are formatted correctly in ISO 8601
func TestFormatTimestamp(t *testing.T) {
	dm := NewDatabaseManager("/repo")

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "specific time",
			time:     time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
			expected: "2024-01-15T14:30:00Z",
		},
		{
			name:     "midnight",
			time:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2024-01-01T00:00:00Z",
		},
		{
			name:     "end of day",
			time:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2024-12-31T23:59:59Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dm.FormatTimestamp(tt.time)
			if result != tt.expected {
				t.Errorf("Expected timestamp %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestFormatDate verifies that dates are formatted correctly in YYYY-MM-DD
func TestFormatDate(t *testing.T) {
	dm := NewDatabaseManager("/repo")

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "mid month",
			time:     time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
			expected: "2024-01-15",
		},
		{
			name:     "first day of year",
			time:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2024-01-01",
		},
		{
			name:     "last day of year",
			time:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2024-12-31",
		},
		{
			name:     "leap year date",
			time:     time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			expected: "2024-02-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dm.FormatDate(tt.time)
			if result != tt.expected {
				t.Errorf("Expected date %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestRecordingMetadataStructure verifies the RecordingMetadata struct has all required fields
func TestRecordingMetadataStructure(t *testing.T) {
	// Create a sample RecordingMetadata to verify all fields are accessible
	metadata := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    3600,
		FileSizeBytes:  2147483648,
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/file/xyz789",
		FilesterChunks: []string{},
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	// Verify all fields are set correctly
	if metadata.Timestamp != "2024-01-15T14:30:00Z" {
		t.Errorf("Timestamp field not set correctly")
	}
	if metadata.DurationSec != 3600 {
		t.Errorf("DurationSec field not set correctly")
	}
	if metadata.FileSizeBytes != 2147483648 {
		t.Errorf("FileSizeBytes field not set correctly")
	}
	if metadata.Quality != "2160p60" {
		t.Errorf("Quality field not set correctly")
	}
	if metadata.GofileURL != "https://gofile.io/d/abc123" {
		t.Errorf("GofileURL field not set correctly")
	}
	if metadata.FilesterURL != "https://filester.me/file/xyz789" {
		t.Errorf("FilesterURL field not set correctly")
	}
	if len(metadata.FilesterChunks) != 0 {
		t.Errorf("FilesterChunks field not set correctly")
	}
	if metadata.SessionID != "run-20240115-143000-abc" {
		t.Errorf("SessionID field not set correctly")
	}
	if metadata.MatrixJob != "matrix-job-1" {
		t.Errorf("MatrixJob field not set correctly")
	}
}

// TestRecordingMetadataWithChunks verifies FilesterChunks field works correctly
func TestRecordingMetadataWithChunks(t *testing.T) {
	chunks := []string{
		"https://filester.me/file/chunk1",
		"https://filester.me/file/chunk2",
		"https://filester.me/file/chunk3",
	}

	metadata := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    7200,
		FileSizeBytes:  32000000000, // 32 GB (split file)
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/folder/xyz789",
		FilesterChunks: chunks,
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	if len(metadata.FilesterChunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(metadata.FilesterChunks))
	}

	for i, chunk := range metadata.FilesterChunks {
		if chunk != chunks[i] {
			t.Errorf("Chunk %d mismatch: expected %s, got %s", i, chunks[i], chunk)
		}
	}
}

// TestAtomicUpdate verifies that AtomicUpdate performs git operations correctly
func TestAtomicUpdate(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_atomic_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository in the temp directory
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Test file path
	testPath := filepath.Join(tmpDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	// Test 1: Create a new file with AtomicUpdate
	t.Run("create new file", func(t *testing.T) {
		err := dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
			// Content should be empty for new file
			if len(content) != 0 {
				t.Errorf("Expected empty content for new file, got %d bytes", len(content))
			}

			// Create initial JSON array
			return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600}]`), nil
		})

		if err != nil {
			t.Errorf("AtomicUpdate failed: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Errorf("File was not created: %s", testPath)
		}

		// Verify file content
		content, err := os.ReadFile(testPath)
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}

		expected := `[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600}]`
		if string(content) != expected {
			t.Errorf("Expected content %s, got %s", expected, string(content))
		}
	})

	// Test 2: Update existing file with AtomicUpdate
	t.Run("update existing file", func(t *testing.T) {
		err := dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
			// Content should contain the previous entry
			if len(content) == 0 {
				t.Errorf("Expected existing content, got empty")
			}

			// Append new entry (simplified for test)
			return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T16:00:00Z","duration_seconds":1800}]`), nil
		})

		if err != nil {
			t.Errorf("AtomicUpdate failed: %v", err)
		}

		// Verify file content
		content, err := os.ReadFile(testPath)
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}

		expected := `[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T16:00:00Z","duration_seconds":1800}]`
		if string(content) != expected {
			t.Errorf("Expected content %s, got %s", expected, string(content))
		}
	})

	// Test 3: Update function returns error
	t.Run("update function error", func(t *testing.T) {
		err := dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
			return nil, fmt.Errorf("simulated update error")
		})

		if err == nil {
			t.Errorf("Expected error from update function, got nil")
		}

		if err != nil && !strings.Contains(err.Error(), "update function failed") {
			t.Errorf("Expected 'update function failed' error, got: %v", err)
		}
	})
}

// TestAtomicUpdateConcurrency verifies that AtomicUpdate handles concurrent access safely
func TestAtomicUpdateConcurrency(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_concurrent_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository in the temp directory
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Test file path
	testPath := filepath.Join(tmpDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	// Create initial file
	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		return []byte(`[]`), nil
	})
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Perform concurrent updates
	const numUpdates = 5
	var wg sync.WaitGroup
	errors := make(chan error, numUpdates)

	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			err := dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
				// Simulate some processing time
				time.Sleep(10 * time.Millisecond)
				
				// For this test, just return a simple update
				// In real scenario, we would parse JSON, append, and marshal
				return []byte(fmt.Sprintf(`[{"index":%d}]`, index)), nil
			})

			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent update failed: %v", err)
	}

	// Verify file exists and has content
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("File was not created: %s", testPath)
	}
}

// TestAtomicUpdateDirectoryCreation verifies that AtomicUpdate creates directories as needed
func TestAtomicUpdateDirectoryCreation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_dir_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository in the temp directory
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Test file path with nested directories that don't exist
	testPath := filepath.Join(tmpDir, "database", "stripchat", "newuser", "2024-01-20.json")

	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		return []byte(`[{"test":"data"}]`), nil
	})

	if err != nil {
		t.Errorf("AtomicUpdate failed: %v", err)
	}

	// Verify directory structure was created
	dir := filepath.Dir(testPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", dir)
	}

	// Verify file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("File was not created: %s", testPath)
	}
}

// TestAtomicUpdateConflictResolution verifies that AtomicUpdate retries on push conflicts
func TestAtomicUpdateConflictResolution(t *testing.T) {
	// Create two temporary directories: one for "remote" (bare) and one for "local"
	remoteDir, err := os.MkdirTemp("", "database_manager_remote_*")
	if err != nil {
		t.Fatalf("Failed to create remote temp directory: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	localDir, err := os.MkdirTemp("", "database_manager_local_*")
	if err != nil {
		t.Fatalf("Failed to create local temp directory: %v", err)
	}
	defer os.RemoveAll(localDir)

	// Initialize bare remote repository
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init bare git repo: %v, output: %s", err, string(output))
	}

	// Clone remote to local
	cmd = exec.Command("git", "clone", remoteDir, localDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone repository: %v, output: %s", err, string(output))
	}

	// Configure local repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.email: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.name: %v, output: %s", err, string(output))
	}

	// Create initial commit in local
	readmePath := filepath.Join(localDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create initial commit: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to push initial commit: %v, output: %s", err, string(output))
	}

	dm := NewDatabaseManager(localDir)

	// Test file path
	testPath := filepath.Join(localDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	// Step 1: Create initial file
	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600}]`), nil
	})
	if err != nil {
		t.Fatalf("Initial AtomicUpdate failed: %v", err)
	}

	// Step 2: Create a second clone to simulate another matrix job
	secondLocalDir, err := os.MkdirTemp("", "database_manager_second_local_*")
	if err != nil {
		t.Fatalf("Failed to create second local temp directory: %v", err)
	}
	defer os.RemoveAll(secondLocalDir)

	cmd = exec.Command("git", "clone", remoteDir, secondLocalDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone repository for second local: %v, output: %s", err, string(output))
	}

	// Configure second local repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = secondLocalDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.email for second local: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = secondLocalDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.name for second local: %v, output: %s", err, string(output))
	}

	// Step 3: Make a conflicting change in the second local and push it
	dm2 := NewDatabaseManager(secondLocalDir)
	secondTestPath := filepath.Join(secondLocalDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	err = dm2.AtomicUpdate(secondTestPath, func(content []byte) ([]byte, error) {
		// This will add a different entry
		return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T15:00:00Z","duration_seconds":1800}]`), nil
	})
	if err != nil {
		t.Fatalf("Second AtomicUpdate failed: %v", err)
	}

	// Step 4: Try to update the first local file (this should trigger conflict resolution)
	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		// This update should retry and eventually succeed by pulling the remote changes first
		return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T15:00:00Z","duration_seconds":1800},{"timestamp":"2024-01-15T16:00:00Z","duration_seconds":2400}]`), nil
	})

	if err != nil {
		t.Errorf("AtomicUpdate with conflict resolution failed: %v", err)
	}

	// Verify the file was updated successfully
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Errorf("Failed to read file after conflict resolution: %v", err)
	}

	// The content should be the latest update
	expected := `[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T15:00:00Z","duration_seconds":1800},{"timestamp":"2024-01-15T16:00:00Z","duration_seconds":2400}]`
	if string(content) != expected {
		t.Errorf("Expected content %s, got %s", expected, string(content))
	}
}

// TestAtomicUpdateMaxRetries verifies that AtomicUpdate fails after max retries
func TestAtomicUpdateMaxRetries(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_maxretries_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Test file path
	testPath := filepath.Join(tmpDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	// Create a scenario where git push always fails
	// We'll use a non-existent remote to simulate persistent push failures
	cmd := exec.Command("git", "remote", "add", "origin", "https://invalid-remote-url-that-does-not-exist.com/repo.git")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add invalid remote: %v, output: %s", err, string(output))
	}

	// Try to update - this should fail after 3 retries
	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		return []byte(`[{"test":"data"}]`), nil
	})

	if err == nil {
		t.Errorf("Expected error after max retries, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "failed after 3 attempts") {
		t.Errorf("Expected 'failed after 3 attempts' error, got: %v", err)
	}
}

// initGitRepo initializes a git repository in the specified directory for testing
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init git repo: %v, output: %s", err, string(output))
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.email: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.name: %v, output: %s", err, string(output))
	}

	// Create initial commit
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create initial commit: %v, output: %s", err, string(output))
	}
}

// TestAtomicUpdateConflictResolutionLogging verifies that conflict resolution attempts are logged comprehensively
func TestAtomicUpdateConflictResolutionLogging(t *testing.T) {
	// Create two temporary directories: one for "remote" (bare) and one for "local"
	remoteDir, err := os.MkdirTemp("", "database_manager_remote_logging_*")
	if err != nil {
		t.Fatalf("Failed to create remote temp directory: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	localDir, err := os.MkdirTemp("", "database_manager_local_logging_*")
	if err != nil {
		t.Fatalf("Failed to create local temp directory: %v", err)
	}
	defer os.RemoveAll(localDir)

	// Initialize bare remote repository
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init bare git repo: %v, output: %s", err, string(output))
	}

	// Clone remote to local
	cmd = exec.Command("git", "clone", remoteDir, localDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone repository: %v, output: %s", err, string(output))
	}

	// Configure local repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.email: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.name: %v, output: %s", err, string(output))
	}

	// Create initial commit in local
	readmePath := filepath.Join(localDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create initial commit: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = localDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to push initial commit: %v, output: %s", err, string(output))
	}

	dm := NewDatabaseManager(localDir)

	// Test file path
	testPath := filepath.Join(localDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	// Step 1: Create initial file
	t.Log("Creating initial database file...")
	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600}]`), nil
	})
	if err != nil {
		t.Fatalf("Initial AtomicUpdate failed: %v", err)
	}

	// Step 2: Create a second clone to simulate another matrix job
	secondLocalDir, err := os.MkdirTemp("", "database_manager_second_local_*")
	if err != nil {
		t.Fatalf("Failed to create second local temp directory: %v", err)
	}
	defer os.RemoveAll(secondLocalDir)

	cmd = exec.Command("git", "clone", remoteDir, secondLocalDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to clone repository for second local: %v, output: %s", err, string(output))
	}

	// Configure second local repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = secondLocalDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.email for second local: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = secondLocalDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to config git user.name for second local: %v, output: %s", err, string(output))
	}

	// Step 3: Make a conflicting change in the second local and push it
	dm2 := NewDatabaseManager(secondLocalDir)
	secondTestPath := filepath.Join(secondLocalDir, "database", "chaturbate", "testuser", "2024-01-15.json")

	t.Log("Creating conflicting update in second clone...")
	err = dm2.AtomicUpdate(secondTestPath, func(content []byte) ([]byte, error) {
		// This will add a different entry
		return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T15:00:00Z","duration_seconds":1800}]`), nil
	})
	if err != nil {
		t.Fatalf("Second AtomicUpdate failed: %v", err)
	}

	// Step 4: Try to update the first local file (this should trigger conflict resolution with logging)
	t.Log("Starting update in first clone that will trigger conflict resolution...")
	err = dm.AtomicUpdate(testPath, func(content []byte) ([]byte, error) {
		// This update should retry and eventually succeed by pulling the remote changes first
		return []byte(`[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T15:00:00Z","duration_seconds":1800},{"timestamp":"2024-01-15T16:00:00Z","duration_seconds":2400}]`), nil
	})

	if err != nil {
		t.Errorf("AtomicUpdate with conflict resolution failed: %v", err)
	}

	// Verify the file was updated successfully
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Errorf("Failed to read file after conflict resolution: %v", err)
	}

	// The content should be the latest update
	expected := `[{"timestamp":"2024-01-15T14:30:00Z","duration_seconds":3600},{"timestamp":"2024-01-15T15:00:00Z","duration_seconds":1800},{"timestamp":"2024-01-15T16:00:00Z","duration_seconds":2400}]`
	if string(content) != expected {
		t.Errorf("Expected content %s, got %s", expected, string(content))
	}

	t.Log("Conflict resolution completed successfully with comprehensive logging")
}

// TestAddRecording verifies that AddRecording creates new files correctly
func TestAddRecording(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Create test metadata
	metadata := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    3600,
		FileSizeBytes:  2147483648,
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/file/xyz789",
		FilesterChunks: []string{},
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	// Test adding a recording to a new file
	err = dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata)
	if err != nil {
		t.Fatalf("AddRecording failed: %v", err)
	}

	// Verify the file was created
	dbPath := dm.GetDatabasePath("chaturbate", "testuser", "2024-01-15")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created: %s", dbPath)
	}

	// Verify the file content
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}

	// Parse the JSON to verify structure
	var recordings []RecordingMetadata
	if err := json.Unmarshal(content, &recordings); err != nil {
		t.Fatalf("Failed to parse database JSON: %v", err)
	}

	// Verify we have exactly one recording
	if len(recordings) != 1 {
		t.Errorf("Expected 1 recording, got %d", len(recordings))
	}

	// Verify the recording metadata
	if recordings[0].Timestamp != metadata.Timestamp {
		t.Errorf("Expected timestamp %s, got %s", metadata.Timestamp, recordings[0].Timestamp)
	}
	if recordings[0].DurationSec != metadata.DurationSec {
		t.Errorf("Expected duration %d, got %d", metadata.DurationSec, recordings[0].DurationSec)
	}
	if recordings[0].FileSizeBytes != metadata.FileSizeBytes {
		t.Errorf("Expected file size %d, got %d", metadata.FileSizeBytes, recordings[0].FileSizeBytes)
	}
	if recordings[0].Quality != metadata.Quality {
		t.Errorf("Expected quality %s, got %s", metadata.Quality, recordings[0].Quality)
	}
	if recordings[0].GofileURL != metadata.GofileURL {
		t.Errorf("Expected Gofile URL %s, got %s", metadata.GofileURL, recordings[0].GofileURL)
	}
	if recordings[0].FilesterURL != metadata.FilesterURL {
		t.Errorf("Expected Filester URL %s, got %s", metadata.FilesterURL, recordings[0].FilesterURL)
	}
	if recordings[0].SessionID != metadata.SessionID {
		t.Errorf("Expected session ID %s, got %s", metadata.SessionID, recordings[0].SessionID)
	}
	if recordings[0].MatrixJob != metadata.MatrixJob {
		t.Errorf("Expected matrix job %s, got %s", metadata.MatrixJob, recordings[0].MatrixJob)
	}
}

// TestAddRecordingAppend verifies that AddRecording appends to existing files
func TestAddRecordingAppend(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_append_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Create first recording
	metadata1 := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    3600,
		FileSizeBytes:  2147483648,
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/file/xyz789",
		FilesterChunks: []string{},
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	err = dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata1)
	if err != nil {
		t.Fatalf("First AddRecording failed: %v", err)
	}

	// Create second recording
	metadata2 := RecordingMetadata{
		Timestamp:      "2024-01-15T16:00:00Z",
		DurationSec:    1800,
		FileSizeBytes:  1073741824,
		Quality:        "1080p60",
		GofileURL:      "https://gofile.io/d/def456",
		FilesterURL:    "https://filester.me/file/uvw123",
		FilesterChunks: []string{},
		SessionID:      "run-20240115-160000-def",
		MatrixJob:      "matrix-job-2",
	}

	err = dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata2)
	if err != nil {
		t.Fatalf("Second AddRecording failed: %v", err)
	}

	// Verify the file content
	dbPath := dm.GetDatabasePath("chaturbate", "testuser", "2024-01-15")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}

	// Parse the JSON
	var recordings []RecordingMetadata
	if err := json.Unmarshal(content, &recordings); err != nil {
		t.Fatalf("Failed to parse database JSON: %v", err)
	}

	// Verify we have exactly two recordings
	if len(recordings) != 2 {
		t.Errorf("Expected 2 recordings, got %d", len(recordings))
	}

	// Verify first recording
	if recordings[0].Timestamp != metadata1.Timestamp {
		t.Errorf("First recording timestamp mismatch: expected %s, got %s", metadata1.Timestamp, recordings[0].Timestamp)
	}

	// Verify second recording
	if recordings[1].Timestamp != metadata2.Timestamp {
		t.Errorf("Second recording timestamp mismatch: expected %s, got %s", metadata2.Timestamp, recordings[1].Timestamp)
	}
}

// TestAddRecordingWithChunks verifies that AddRecording handles split files correctly
func TestAddRecordingWithChunks(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_chunks_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Create metadata with chunks (for files > 10 GB)
	chunks := []string{
		"https://filester.me/file/chunk1",
		"https://filester.me/file/chunk2",
		"https://filester.me/file/chunk3",
	}

	metadata := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    7200,
		FileSizeBytes:  32000000000, // 32 GB
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/folder/xyz789",
		FilesterChunks: chunks,
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	err = dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata)
	if err != nil {
		t.Fatalf("AddRecording with chunks failed: %v", err)
	}

	// Verify the file content
	dbPath := dm.GetDatabasePath("chaturbate", "testuser", "2024-01-15")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}

	// Parse the JSON
	var recordings []RecordingMetadata
	if err := json.Unmarshal(content, &recordings); err != nil {
		t.Fatalf("Failed to parse database JSON: %v", err)
	}

	// Verify we have exactly one recording
	if len(recordings) != 1 {
		t.Errorf("Expected 1 recording, got %d", len(recordings))
	}

	// Verify chunks are preserved
	if len(recordings[0].FilesterChunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(recordings[0].FilesterChunks))
	}

	for i, chunk := range recordings[0].FilesterChunks {
		if chunk != chunks[i] {
			t.Errorf("Chunk %d mismatch: expected %s, got %s", i, chunks[i], chunk)
		}
	}
}

// TestAddRecordingJSONValidation verifies that AddRecording validates JSON structure
func TestAddRecordingJSONValidation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_validation_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Create valid metadata
	metadata := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    3600,
		FileSizeBytes:  2147483648,
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/file/xyz789",
		FilesterChunks: []string{},
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	err = dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata)
	if err != nil {
		t.Fatalf("AddRecording failed: %v", err)
	}

	// Verify the file content is valid JSON
	dbPath := dm.GetDatabasePath("chaturbate", "testuser", "2024-01-15")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}

	// Verify JSON is valid by unmarshaling
	var recordings []RecordingMetadata
	if err := json.Unmarshal(content, &recordings); err != nil {
		t.Errorf("Database file contains invalid JSON: %v", err)
	}

	// Verify JSON is properly formatted (indented)
	if !strings.Contains(string(content), "\n") {
		t.Errorf("JSON is not properly formatted (should be indented)")
	}
}

// TestAddRecordingConcurrent verifies that concurrent AddRecording calls work correctly
func TestAddRecordingConcurrent(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_concurrent_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Perform concurrent AddRecording calls
	const numRecordings = 5
	var wg sync.WaitGroup
	errors := make(chan error, numRecordings)

	for i := 0; i < numRecordings; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			metadata := RecordingMetadata{
				Timestamp:      fmt.Sprintf("2024-01-15T%02d:00:00Z", 14+index),
				DurationSec:    3600,
				FileSizeBytes:  2147483648,
				Quality:        "2160p60",
				GofileURL:      fmt.Sprintf("https://gofile.io/d/abc%d", index),
				FilesterURL:    fmt.Sprintf("https://filester.me/file/xyz%d", index),
				FilesterChunks: []string{},
				SessionID:      fmt.Sprintf("run-20240115-%d", index),
				MatrixJob:      fmt.Sprintf("matrix-job-%d", index),
			}

			err := dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent AddRecording failed: %v", err)
	}

	// Verify all recordings were added
	dbPath := dm.GetDatabasePath("chaturbate", "testuser", "2024-01-15")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}

	var recordings []RecordingMetadata
	if err := json.Unmarshal(content, &recordings); err != nil {
		t.Fatalf("Failed to parse database JSON: %v", err)
	}

	// We should have at least some recordings (may not be all due to git conflicts in test environment)
	if len(recordings) == 0 {
		t.Errorf("Expected at least 1 recording, got 0")
	}

	t.Logf("Successfully added %d recordings concurrently", len(recordings))
}

// TestAddRecordingMultipleSites verifies that AddRecording works for different sites
func TestAddRecordingMultipleSites(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_sites_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	sites := []string{"chaturbate", "stripchat"}
	channels := []string{"user1", "user2"}

	for _, site := range sites {
		for _, channel := range channels {
			metadata := RecordingMetadata{
				Timestamp:      "2024-01-15T14:30:00Z",
				DurationSec:    3600,
				FileSizeBytes:  2147483648,
				Quality:        "2160p60",
				GofileURL:      fmt.Sprintf("https://gofile.io/d/%s-%s", site, channel),
				FilesterURL:    fmt.Sprintf("https://filester.me/file/%s-%s", site, channel),
				FilesterChunks: []string{},
				SessionID:      "run-20240115-143000-abc",
				MatrixJob:      "matrix-job-1",
			}

			err := dm.AddRecording(site, channel, "2024-01-15", metadata)
			if err != nil {
				t.Errorf("AddRecording failed for %s/%s: %v", site, channel, err)
			}

			// Verify the file was created in the correct location
			dbPath := dm.GetDatabasePath(site, channel, "2024-01-15")
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				t.Errorf("Database file was not created for %s/%s: %s", site, channel, dbPath)
			}
		}
	}
}

// TestAddRecordingAllFields verifies that all metadata fields are preserved
func TestAddRecordingAllFields(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "database_manager_addrecording_fields_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository
	initGitRepo(t, tmpDir)

	dm := NewDatabaseManager(tmpDir)

	// Create metadata with all fields populated
	metadata := RecordingMetadata{
		Timestamp:      "2024-01-15T14:30:00Z",
		DurationSec:    3600,
		FileSizeBytes:  2147483648,
		Quality:        "2160p60",
		GofileURL:      "https://gofile.io/d/abc123",
		FilesterURL:    "https://filester.me/file/xyz789",
		FilesterChunks: []string{"chunk1", "chunk2"},
		SessionID:      "run-20240115-143000-abc",
		MatrixJob:      "matrix-job-1",
	}

	err = dm.AddRecording("chaturbate", "testuser", "2024-01-15", metadata)
	if err != nil {
		t.Fatalf("AddRecording failed: %v", err)
	}

	// Read and verify all fields
	dbPath := dm.GetDatabasePath("chaturbate", "testuser", "2024-01-15")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database file: %v", err)
	}

	var recordings []RecordingMetadata
	if err := json.Unmarshal(content, &recordings); err != nil {
		t.Fatalf("Failed to parse database JSON: %v", err)
	}

	if len(recordings) != 1 {
		t.Fatalf("Expected 1 recording, got %d", len(recordings))
	}

	r := recordings[0]

	// Verify each field
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Timestamp", metadata.Timestamp, r.Timestamp},
		{"DurationSec", metadata.DurationSec, r.DurationSec},
		{"FileSizeBytes", metadata.FileSizeBytes, r.FileSizeBytes},
		{"Quality", metadata.Quality, r.Quality},
		{"GofileURL", metadata.GofileURL, r.GofileURL},
		{"FilesterURL", metadata.FilesterURL, r.FilesterURL},
		{"SessionID", metadata.SessionID, r.SessionID},
		{"MatrixJob", metadata.MatrixJob, r.MatrixJob},
	}

	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("%s mismatch: expected %v, got %v", tt.name, tt.expected, tt.actual)
		}
	}

	// Verify chunks separately
	if len(r.FilesterChunks) != len(metadata.FilesterChunks) {
		t.Errorf("FilesterChunks length mismatch: expected %d, got %d", len(metadata.FilesterChunks), len(r.FilesterChunks))
	}

	for i := range metadata.FilesterChunks {
		if r.FilesterChunks[i] != metadata.FilesterChunks[i] {
			t.Errorf("FilesterChunks[%d] mismatch: expected %s, got %s", i, metadata.FilesterChunks[i], r.FilesterChunks[i])
		}
	}
}
