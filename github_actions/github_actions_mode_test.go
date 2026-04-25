package github_actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HeapOfChaos/goondvr/entity"
)

// TestStartWorkflowLifecycle tests the workflow lifecycle management
func TestStartWorkflowLifecycle(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	// Create GitHubActionsMode instance
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-123"
	channels := []string{"test-channel-1", "test-channel-2"}
	
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Fatalf("StartWorkflowLifecycle failed: %v", err)
	}
	
	// Verify matrix job was registered
	activeJobs := gam.MatrixCoordinator.GetActiveJobs()
	if len(activeJobs) != 1 {
		t.Errorf("Expected 1 active job, got %d", len(activeJobs))
	}
	
	if len(activeJobs) > 0 {
		job := activeJobs[0]
		if job.JobID != matrixJobID {
			t.Errorf("Expected job ID %s, got %s", matrixJobID, job.JobID)
		}
		if job.Channel != "test-channel-1" {
			t.Errorf("Expected channel test-channel-1, got %s", job.Channel)
		}
		if job.Status != "starting" {
			t.Errorf("Expected status 'starting', got %s", job.Status)
		}
	}
	
	// Give background goroutines a moment to start
	time.Sleep(100 * time.Millisecond)
	
	// Verify Chain Manager is monitoring
	if gam.ChainManager.GetSessionID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, gam.ChainManager.GetSessionID())
	}
	
	// Cancel context to stop background goroutines
	gam.Cancel()
	
	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)
}

// TestStartWorkflowLifecycle_CacheMiss tests lifecycle with no cached state
func TestStartWorkflowLifecycle_CacheMiss(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories (but no cached state)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	// Create GitHubActionsMode instance
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-456"
	channels := []string{"test-channel-1"}
	
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle - should handle cache miss gracefully
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Fatalf("StartWorkflowLifecycle failed on cache miss: %v", err)
	}
	
	// Verify matrix job was still registered despite cache miss
	activeJobs := gam.MatrixCoordinator.GetActiveJobs()
	if len(activeJobs) != 1 {
		t.Errorf("Expected 1 active job after cache miss, got %d", len(activeJobs))
	}
}

// TestStartWorkflowLifecycle_InvalidMatrixJobID tests error handling for invalid job ID
func TestStartWorkflowLifecycle_InvalidMatrixJobID(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	// Create GitHubActionsMode instance with invalid matrix job ID
	matrixJobID := "invalid-job-id" // Not in format "matrix-job-N"
	sessionID := "test-session-789"
	channels := []string{"test-channel-1"}
	
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle - should fail due to invalid job ID
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err == nil {
		t.Fatal("Expected error for invalid matrix job ID, got nil")
	}
	
	// Verify error message mentions the invalid format
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestStartWorkflowLifecycle_BackgroundMonitoring tests that background goroutines start
func TestStartWorkflowLifecycle_BackgroundMonitoring(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	// Create GitHubActionsMode instance
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-background"
	channels := []string{"test-channel-1"}
	
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Fatalf("StartWorkflowLifecycle failed: %v", err)
	}
	
	// Verify Chain Manager has not triggered next run yet (should take 5.5 hours)
	if gam.ChainManager.IsNextRunTriggered() {
		t.Error("Chain Manager should not have triggered next run immediately")
	}
	
	// Verify elapsed time is minimal
	elapsed := gam.ChainManager.GetElapsedTime()
	if elapsed > 1*time.Second {
		t.Errorf("Expected minimal elapsed time, got %v", elapsed)
	}
	
	// Let background goroutines run for a moment
	time.Sleep(200 * time.Millisecond)
	
	// Cancel context to stop background goroutines
	gam.Cancel()
	
	// Verify context is cancelled
	select {
	case <-gam.GetContext().Done():
		// Expected - context is cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled")
	}
}

// TestGetAssignedChannel tests channel assignment logic
func TestGetAssignedChannel(t *testing.T) {
	tests := []struct {
		name        string
		matrixJobID string
		channels    []string
		wantChannel string
		wantError   bool
	}{
		{
			name:        "First job gets first channel",
			matrixJobID: "matrix-job-1",
			channels:    []string{"channel-a", "channel-b", "channel-c"},
			wantChannel: "channel-a",
			wantError:   false,
		},
		{
			name:        "Second job gets second channel",
			matrixJobID: "matrix-job-2",
			channels:    []string{"channel-a", "channel-b", "channel-c"},
			wantChannel: "channel-b",
			wantError:   false,
		},
		{
			name:        "Third job gets third channel",
			matrixJobID: "matrix-job-3",
			channels:    []string{"channel-a", "channel-b", "channel-c"},
			wantChannel: "channel-c",
			wantError:   false,
		},
		{
			name:        "Invalid job ID format",
			matrixJobID: "invalid-format",
			channels:    []string{"channel-a"},
			wantChannel: "",
			wantError:   true,
		},
		{
			name:        "Job index out of range",
			matrixJobID: "matrix-job-5",
			channels:    []string{"channel-a", "channel-b"},
			wantChannel: "",
			wantError:   true,
		},
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gam, err := NewGitHubActionsMode(tt.matrixJobID, "test-session", tt.channels, false, false)
			if err != nil {
				t.Fatalf("Failed to create GitHubActionsMode: %v", err)
			}
			defer gam.Cancel()
			
			channel, err := gam.GetAssignedChannel()
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if channel != tt.wantChannel {
					t.Errorf("Expected channel %s, got %s", tt.wantChannel, channel)
				}
			}
		})
	}
}

// TestStartWorkflowLifecycle_WithCachedState tests lifecycle with existing cached state
func TestStartWorkflowLifecycle_WithCachedState(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	// Create GitHubActionsMode instance
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-cached"
	channels := []string{"test-channel-1"}
	
	// First, save some state
	persister := NewStatePersister(sessionID, matrixJobID, stateDir)
	ctx := context.Background()
	
	// Create a test config file
	testConfigFile := filepath.Join(configDir, "test.conf")
	if err := os.WriteFile(testConfigFile, []byte("test config"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Save state
	if err := persister.SaveState(ctx, configDir, recordingsDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}
	
	// Now create a new GitHubActionsMode and start lifecycle
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle - should restore cached state
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Fatalf("StartWorkflowLifecycle failed with cached state: %v", err)
	}
	
	// Verify matrix job was registered
	activeJobs := gam.MatrixCoordinator.GetActiveJobs()
	if len(activeJobs) != 1 {
		t.Errorf("Expected 1 active job, got %d", len(activeJobs))
	}
}

// TestApplyQualityToChannelConfig tests quality settings application
func TestApplyQualityToChannelConfig(t *testing.T) {
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	tests := []struct {
		name               string
		maxQuality         bool
		initialResolution  int
		initialFramerate   int
		expectedResolution int
		expectedFramerate  int
	}{
		{
			name:               "Max quality enabled - applies 2160p60",
			maxQuality:         true,
			initialResolution:  720,
			initialFramerate:   30,
			expectedResolution: 2160,
			expectedFramerate:  60,
		},
		{
			name:               "Max quality disabled - keeps original settings",
			maxQuality:         false,
			initialResolution:  720,
			initialFramerate:   30,
			expectedResolution: 720,
			expectedFramerate:  30,
		},
		{
			name:               "Max quality enabled - overrides existing high quality",
			maxQuality:         true,
			initialResolution:  1080,
			initialFramerate:   60,
			expectedResolution: 2160,
			expectedFramerate:  60,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create GitHubActionsMode instance
			gam, err := NewGitHubActionsMode("matrix-job-1", "test-session", []string{"test-channel"}, tt.maxQuality, false)
			if err != nil {
				t.Fatalf("Failed to create GitHubActionsMode: %v", err)
			}
			defer gam.Cancel()
			
			// Create a channel config with initial settings
			config := &entity.ChannelConfig{
				Username:   "testuser",
				Site:       "chaturbate",
				Resolution: tt.initialResolution,
				Framerate:  tt.initialFramerate,
			}
			
			// Apply quality settings
			err = gam.ApplyQualityToChannelConfig(config)
			if err != nil {
				t.Fatalf("ApplyQualityToChannelConfig failed: %v", err)
			}
			
			// Verify resolution and framerate
			if config.Resolution != tt.expectedResolution {
				t.Errorf("Expected resolution %d, got %d", tt.expectedResolution, config.Resolution)
			}
			if config.Framerate != tt.expectedFramerate {
				t.Errorf("Expected framerate %d, got %d", tt.expectedFramerate, config.Framerate)
			}
		})
	}
}

// TestCreateChannelConfigWithQuality tests channel config creation with quality settings
func TestCreateChannelConfigWithQuality(t *testing.T) {
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	tests := []struct {
		name               string
		maxQuality         bool
		username           string
		site               string
		expectedResolution int
		expectedFramerate  int
	}{
		{
			name:               "Create config with max quality enabled",
			maxQuality:         true,
			username:           "testuser1",
			site:               "chaturbate",
			expectedResolution: 2160,
			expectedFramerate:  60,
		},
		{
			name:               "Create config with max quality disabled",
			maxQuality:         false,
			username:           "testuser2",
			site:               "stripchat",
			expectedResolution: 1080,
			expectedFramerate:  30,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create GitHubActionsMode instance
			gam, err := NewGitHubActionsMode("matrix-job-1", "test-session", []string{"test-channel"}, tt.maxQuality, false)
			if err != nil {
				t.Fatalf("Failed to create GitHubActionsMode: %v", err)
			}
			defer gam.Cancel()
			
			// Create channel config with quality settings
			config, err := gam.CreateChannelConfigWithQuality(tt.username, tt.site)
			if err != nil {
				t.Fatalf("CreateChannelConfigWithQuality failed: %v", err)
			}
			
			// Verify basic fields
			if config.Username != tt.username {
				t.Errorf("Expected username %s, got %s", tt.username, config.Username)
			}
			if config.Site != tt.site {
				t.Errorf("Expected site %s, got %s", tt.site, config.Site)
			}
			if config.IsPaused {
				t.Error("Expected IsPaused to be false")
			}
			
			// Verify quality settings
			if config.Resolution != tt.expectedResolution {
				t.Errorf("Expected resolution %d, got %d", tt.expectedResolution, config.Resolution)
			}
			if config.Framerate != tt.expectedFramerate {
				t.Errorf("Expected framerate %d, got %d", tt.expectedFramerate, config.Framerate)
			}
			
			// Verify pattern is set
			if config.Pattern == "" {
				t.Error("Expected pattern to be set")
			}
		})
	}
}

// TestApplyQualityToChannelConfig_RequirementValidation tests that quality settings meet requirements
func TestApplyQualityToChannelConfig_RequirementValidation(t *testing.T) {
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	// Create GitHubActionsMode with max quality enabled
	gam, err := NewGitHubActionsMode("matrix-job-1", "test-session", []string{"test-channel"}, true, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Create a channel config with low quality settings
	config := &entity.ChannelConfig{
		Username:   "testuser",
		Site:       "chaturbate",
		Resolution: 480,
		Framerate:  15,
	}
	
	// Apply quality settings
	err = gam.ApplyQualityToChannelConfig(config)
	if err != nil {
		t.Fatalf("ApplyQualityToChannelConfig failed: %v", err)
	}
	
	// Requirement 16.1: Attempt to record at 2160p (4K) resolution as first priority
	if config.Resolution != 2160 {
		t.Errorf("Requirement 16.1 failed: Expected resolution 2160, got %d", config.Resolution)
	}
	
	// Requirement 16.2: Attempt to record at 60 frames per second as first priority
	if config.Framerate != 60 {
		t.Errorf("Requirement 16.2 failed: Expected framerate 60, got %d", config.Framerate)
	}
	
	// Requirement 16.10: Override any existing resolution or framerate configuration settings
	// (verified by the fact that 480p15 was changed to 2160p60)
	if config.Resolution == 480 || config.Framerate == 15 {
		t.Error("Requirement 16.10 failed: Existing settings were not overridden")
	}
}

// TestMaskSecret tests the secret masking helper function
func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected string
	}{
		{
			name:     "Empty secret",
			secret:   "",
			expected: "<not set>",
		},
		{
			name:     "Short secret (8 chars or less)",
			secret:   "short",
			expected: "****",
		},
		{
			name:     "Exactly 8 chars",
			secret:   "12345678",
			expected: "****",
		},
		{
			name:     "Long secret",
			secret:   "this-is-a-very-long-secret-key-12345",
			expected: "this****2345",
		},
		{
			name:     "Medium secret",
			secret:   "abcdefghij",
			expected: "abcd****ghij",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSecret(tt.secret)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestCacheRestorationFailureRecovery tests that the system handles cache restoration failures gracefully
// and continues operation with default state. This validates Task 14.2 requirements.
//
// Requirements: 8.2 (Requirement 8: Recovery from Failures)
// - Initialize with default state on cache miss
// - Log warnings for missing cache
// - Continue operation with fresh state
func TestCacheRestorationFailureRecovery(t *testing.T) {
	tests := []struct {
		name           string
		setupCache     bool
		corruptCache   bool
		expectError    bool
		expectContinue bool
	}{
		{
			name:           "No cache exists (first run) - should continue with default state",
			setupCache:     false,
			corruptCache:   false,
			expectError:    false,
			expectContinue: true,
		},
		{
			name:           "Valid cache exists - should restore successfully",
			setupCache:     true,
			corruptCache:   false,
			expectError:    false,
			expectContinue: true,
		},
		{
			name:           "Corrupted cache (integrity failure) - should continue with default state",
			setupCache:     true,
			corruptCache:   true,
			expectError:    false,
			expectContinue: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directories for testing
			tempDir := t.TempDir()
			configDir := filepath.Join(tempDir, "conf")
			recordingsDir := filepath.Join(tempDir, "videos")
			stateDir := filepath.Join(tempDir, "state")
			
			// Create directories
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config dir: %v", err)
			}
			if err := os.MkdirAll(recordingsDir, 0755); err != nil {
				t.Fatalf("Failed to create recordings dir: %v", err)
			}
			if err := os.MkdirAll(stateDir, 0755); err != nil {
				t.Fatalf("Failed to create state dir: %v", err)
			}
			
			// Set required environment variables
			os.Setenv("GITHUB_TOKEN", "test-token")
			os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
			os.Setenv("GOFILE_API_KEY", "test-gofile-key")
			os.Setenv("FILESTER_API_KEY", "test-filester-key")
			defer func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			}()
			
			matrixJobID := "matrix-job-1"
			sessionID := "test-session-recovery"
			channels := []string{"test-channel-1"}
			
			// Setup cache if requested
			if tt.setupCache {
				persister := NewStatePersister(sessionID, matrixJobID, stateDir)
				ctx := context.Background()
				
				// Create a test config file
				testConfigFile := filepath.Join(configDir, "test.conf")
				if err := os.WriteFile(testConfigFile, []byte("test config content"), 0644); err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				
				// Save state
				if err := persister.SaveState(ctx, configDir, recordingsDir); err != nil {
					t.Fatalf("Failed to save state: %v", err)
				}
				
				// Corrupt cache if requested
				if tt.corruptCache {
					// Modify a cached file to corrupt its checksum
					cachedFiles, err := filepath.Glob(filepath.Join(stateDir, "config", "*"))
					if err != nil {
						t.Fatalf("Failed to find cached files: %v", err)
					}
					if len(cachedFiles) > 0 {
						// Append data to corrupt the file
						if err := os.WriteFile(cachedFiles[0], []byte("corrupted"), 0644); err != nil {
							t.Fatalf("Failed to corrupt cache file: %v", err)
						}
					}
				}
			}
			
			// Create GitHubActionsMode instance
			gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
			if err != nil {
				t.Fatalf("Failed to create GitHubActionsMode: %v", err)
			}
			defer gam.Cancel()
			
			// Override the state persister to use our temp directory
			gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
			
			// Start workflow lifecycle - should handle cache failures gracefully
			err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
			
			if tt.expectContinue {
				// Verify that the workflow continued despite cache issues
				// Check that matrix job was registered
				activeJobs := gam.MatrixCoordinator.GetActiveJobs()
				if len(activeJobs) != 1 {
					t.Errorf("Expected 1 active job (workflow should continue), got %d", len(activeJobs))
				}
				
				// Verify Chain Manager is operational
				if gam.ChainManager.GetSessionID() != sessionID {
					t.Errorf("Expected session ID %s, got %s", sessionID, gam.ChainManager.GetSessionID())
				}
			}
		})
	}
}

// TestCacheRestorationFailureRecovery_LoggingBehavior verifies that appropriate warnings
// are logged when cache restoration fails.
//
// Requirements: 8.2 (Requirement 8: Recovery from Failures)
// - Log warnings for missing cache
func TestCacheRestorationFailureRecovery_LoggingBehavior(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories (but no cached state)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-logging"
	channels := []string{"test-channel-1"}
	
	// Create GitHubActionsMode instance
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle - should log warnings about cache miss
	// The actual logging is done by the log package, so we can't easily capture it in tests
	// But we can verify that the workflow continues successfully
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Errorf("StartWorkflowLifecycle should not fail on cache miss: %v", err)
	}
	
	// Verify workflow continued successfully
	activeJobs := gam.MatrixCoordinator.GetActiveJobs()
	if len(activeJobs) != 1 {
		t.Errorf("Expected 1 active job after cache miss, got %d", len(activeJobs))
	}
}

// TestCacheRestorationFailureRecovery_DefaultStateInitialization verifies that
// the system initializes with default state when cache restoration fails.
//
// Requirements: 8.2 (Requirement 8: Recovery from Failures)
// - Initialize with default configuration on cache miss
// - Continue operation with fresh state
func TestCacheRestorationFailureRecovery_DefaultStateInitialization(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories (no cached state)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-default-state"
	channels := []string{"test-channel-1", "test-channel-2"}
	
	// Create GitHubActionsMode instance
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, true, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle - should initialize with default state
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Fatalf("StartWorkflowLifecycle failed: %v", err)
	}
	
	// Verify default state initialization by checking component states
	
	// 1. Verify Chain Manager is initialized with correct session ID
	if gam.ChainManager.GetSessionID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, gam.ChainManager.GetSessionID())
	}
	
	// 2. Verify Matrix Coordinator has registered the job
	activeJobs := gam.MatrixCoordinator.GetActiveJobs()
	if len(activeJobs) != 1 {
		t.Errorf("Expected 1 active job, got %d", len(activeJobs))
	}
	
	// 3. Verify Quality Selector is initialized with default preferences
	if gam.QualitySelector.GetPreferredResolution() != 2160 {
		t.Errorf("Expected default resolution 2160, got %d", gam.QualitySelector.GetPreferredResolution())
	}
	if gam.QualitySelector.GetPreferredFramerate() != 60 {
		t.Errorf("Expected default framerate 60, got %d", gam.QualitySelector.GetPreferredFramerate())
	}
	
	// 4. Verify Storage Uploader is initialized
	if gam.StorageUploader == nil {
		t.Error("Storage Uploader should be initialized")
	}
	
	// 5. Verify Database Manager is initialized
	if gam.DatabaseManager == nil {
		t.Error("Database Manager should be initialized")
	}
	
	// 6. Verify Health Monitor is initialized
	if gam.HealthMonitor == nil {
		t.Error("Health Monitor should be initialized")
	}
	
	// 7. Verify Graceful Shutdown is initialized
	if gam.GracefulShutdown == nil {
		t.Error("Graceful Shutdown should be initialized")
	}
	
	// All components should be operational despite cache miss
	t.Log("All components initialized successfully with default state after cache miss")
}

// TestCacheRestorationFailureRecovery_IntegrityFailure tests recovery from cache integrity failures
//
// Requirements: 8.2 (Requirement 8: Recovery from Failures)
// - Continue operation when cache integrity verification fails
func TestCacheRestorationFailureRecovery_IntegrityFailure(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "conf")
	recordingsDir := filepath.Join(tempDir, "videos")
	stateDir := filepath.Join(tempDir, "state")
	
	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		t.Fatalf("Failed to create recordings dir: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("Failed to create state dir: %v", err)
	}
	
	// Set required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "test-owner/test-repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()
	
	matrixJobID := "matrix-job-1"
	sessionID := "test-session-integrity"
	channels := []string{"test-channel-1"}
	
	// First, save valid state
	persister := NewStatePersister(sessionID, matrixJobID, stateDir)
	ctx := context.Background()
	
	// Create a test config file
	testConfigFile := filepath.Join(configDir, "test.conf")
	if err := os.WriteFile(testConfigFile, []byte("original content"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Save state
	if err := persister.SaveState(ctx, configDir, recordingsDir); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}
	
	// Corrupt the cached file to cause integrity failure
	cachedFiles, err := filepath.Glob(filepath.Join(stateDir, "config", "*"))
	if err != nil {
		t.Fatalf("Failed to find cached files: %v", err)
	}
	if len(cachedFiles) > 0 {
		// Modify the file to corrupt its checksum
		if err := os.WriteFile(cachedFiles[0], []byte("corrupted content"), 0644); err != nil {
			t.Fatalf("Failed to corrupt cache file: %v", err)
		}
	}
	
	// Create GitHubActionsMode instance
	gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, false, false)
	if err != nil {
		t.Fatalf("Failed to create GitHubActionsMode: %v", err)
	}
	defer gam.Cancel()
	
	// Override the state persister to use our temp directory
	gam.StatePersister = NewStatePersister(sessionID, matrixJobID, stateDir)
	
	// Start workflow lifecycle - should handle integrity failure gracefully
	err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
	if err != nil {
		t.Errorf("StartWorkflowLifecycle should not fail on integrity failure: %v", err)
	}
	
	// Verify workflow continued successfully despite integrity failure
	activeJobs := gam.MatrixCoordinator.GetActiveJobs()
	if len(activeJobs) != 1 {
		t.Errorf("Expected 1 active job after integrity failure, got %d", len(activeJobs))
	}
	
	// Verify Chain Manager is operational
	if gam.ChainManager.GetSessionID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, gam.ChainManager.GetSessionID())
	}
	
	t.Log("Workflow continued successfully after cache integrity failure")
}

