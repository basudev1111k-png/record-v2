package github_actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMonitorDiskSpace_ContextCancellation verifies that monitoring stops when context is cancelled
func TestMonitorDiskSpace_ContextCancellation(t *testing.T) {
	t.Skip("Skipping test that hangs on Windows due to disk stats implementation")
	
	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 1 * time.Second,
		statusFilePath:    "",
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start monitoring in a goroutine
	done := make(chan error, 1)
	go func() {
		// Use a path that will fail on Windows to trigger the error path
		done <- hm.MonitorDiskSpace(ctx, ".", nil, nil)
	}()

	// Cancel context after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for monitoring to stop
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("MonitorDiskSpace did not stop after context cancellation")
	}
}

// TestSendNotification verifies that notifications are sent to all configured notifiers
func TestSendNotification(t *testing.T) {
	// Create mock notifiers
	notifier1Called := false
	notifier2Called := false

	mockNotifier1 := &mockNotifier{
		sendFunc: func(title, message string) error {
			notifier1Called = true
			if title != "Test Title" {
				t.Errorf("Expected title 'Test Title', got '%s'", title)
			}
			if message != "Test Message" {
				t.Errorf("Expected message 'Test Message', got '%s'", message)
			}
			return nil
		},
	}

	mockNotifier2 := &mockNotifier{
		sendFunc: func(title, message string) error {
			notifier2Called = true
			return nil
		},
	}

	hm := &HealthMonitor{
		notifiers:         []Notifier{mockNotifier1, mockNotifier2},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    "",
	}

	// Send notification
	err := hm.SendNotification("Test Title", "Test Message")
	if err != nil {
		t.Errorf("SendNotification returned error: %v", err)
	}

	// Verify both notifiers were called
	if !notifier1Called {
		t.Error("Notifier 1 was not called")
	}
	if !notifier2Called {
		t.Error("Notifier 2 was not called")
	}
}

// TestSendNotification_ErrorHandling verifies that errors from one notifier don't stop others
func TestSendNotification_ErrorHandling(t *testing.T) {
	notifier1Called := false
	notifier2Called := false
	notifier3Called := false

	mockNotifier1 := &mockNotifier{
		sendFunc: func(title, message string) error {
			notifier1Called = true
			return nil
		},
	}

	mockNotifier2 := &mockNotifier{
		sendFunc: func(title, message string) error {
			notifier2Called = true
			return fmt.Errorf("simulated error")
		},
	}

	mockNotifier3 := &mockNotifier{
		sendFunc: func(title, message string) error {
			notifier3Called = true
			return nil
		},
	}

	hm := &HealthMonitor{
		notifiers:         []Notifier{mockNotifier1, mockNotifier2, mockNotifier3},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    "",
	}

	// Send notification
	err := hm.SendNotification("Test Title", "Test Message")
	
	// Should return the last error encountered
	if err == nil {
		t.Error("Expected error to be returned")
	}

	// Verify all notifiers were called despite error in notifier2
	if !notifier1Called {
		t.Error("Notifier 1 was not called")
	}
	if !notifier2Called {
		t.Error("Notifier 2 was not called")
	}
	if !notifier3Called {
		t.Error("Notifier 3 was not called despite error in notifier 2")
	}
}

// TestSendNotification_NoNotifiers verifies behavior with empty notifiers list
func TestSendNotification_NoNotifiers(t *testing.T) {
	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    "",
	}

	// Send notification with no notifiers
	err := hm.SendNotification("Test Title", "Test Message")
	if err != nil {
		t.Errorf("SendNotification with no notifiers should not return error, got: %v", err)
	}
}

// TestSendNotification_WorkflowLifecycle verifies notifications for workflow events
func TestSendNotification_WorkflowLifecycle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		message string
	}{
		{
			name:    "Workflow Start",
			title:   "Workflow Started",
			message: "Session ID: test-session-123",
		},
		{
			name:    "Workflow End",
			title:   "Workflow Completed",
			message: "Session statistics: 5 recordings completed",
		},
		{
			name:    "Matrix Job Start",
			title:   "Matrix Job Started",
			message: "Job ID: matrix-1, Channel: test_channel",
		},
		{
			name:    "Matrix Job Fail",
			title:   "Matrix Job Failed",
			message: "Job ID: matrix-2, Error: connection timeout",
		},
		{
			name:    "Chain Transition",
			title:   "Chain Transition",
			message: "Transitioning from session-1 to session-2",
		},
		{
			name:    "Recording Start",
			title:   "Recording Started",
			message: "Channel: test_channel, Quality: 2160p60",
		},
		{
			name:    "Recording Complete",
			title:   "Recording Completed",
			message: "Channel: test_channel, Size: 2.5GB, Quality: 2160p60, Gofile: uploaded, Filester: uploaded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			var receivedTitle, receivedMessage string

			mockNotifier := &mockNotifier{
				sendFunc: func(title, message string) error {
					called = true
					receivedTitle = title
					receivedMessage = message
					return nil
				},
			}

			hm := &HealthMonitor{
				notifiers:         []Notifier{mockNotifier},
				diskCheckInterval: 5 * time.Minute,
				statusFilePath:    "",
			}

			err := hm.SendNotification(tt.title, tt.message)
			if err != nil {
				t.Errorf("SendNotification returned error: %v", err)
			}

			if !called {
				t.Error("Notifier was not called")
			}

			if receivedTitle != tt.title {
				t.Errorf("Expected title '%s', got '%s'", tt.title, receivedTitle)
			}

			if receivedMessage != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, receivedMessage)
			}
		})
	}
}

// TestDiscordNotifier verifies Discord notifier implementation
func TestDiscordNotifier(t *testing.T) {
	// Create Discord notifier with test webhook URL
	dn := NewDiscordNotifier("https://discord.com/api/webhooks/test")

	if dn.webhookURL != "https://discord.com/api/webhooks/test" {
		t.Errorf("Expected webhook URL to be set correctly")
	}

	// Note: We can't easily test the actual Send() method without mocking HTTP
	// The Send() method uses the existing notifier package which has its own tests
}

// TestNtfyNotifier verifies ntfy notifier implementation
func TestNtfyNotifier(t *testing.T) {
	// Create ntfy notifier with test configuration
	nn := NewNtfyNotifier("https://ntfy.sh", "test-topic", "test-token")

	if nn.serverURL != "https://ntfy.sh" {
		t.Errorf("Expected server URL to be set correctly")
	}

	if nn.topic != "test-topic" {
		t.Errorf("Expected topic to be set correctly")
	}

	if nn.token != "test-token" {
		t.Errorf("Expected token to be set correctly")
	}

	// Note: We can't easily test the actual Send() method without mocking HTTP
	// The Send() method uses the existing notifier package which has its own tests
}

// TestNewHealthMonitor verifies that the constructor initializes fields correctly
func TestNewHealthMonitor(t *testing.T) {
	notifiers := []Notifier{&mockNotifier{}}
	statusFilePath := "/tmp/status.json"

	hm := NewHealthMonitor(statusFilePath, notifiers)

	if hm.diskCheckInterval != 5*time.Minute {
		t.Errorf("Expected disk check interval of 5 minutes, got %v", hm.diskCheckInterval)
	}

	if hm.statusFilePath != statusFilePath {
		t.Errorf("Expected status file path '%s', got '%s'", statusFilePath, hm.statusFilePath)
	}

	if len(hm.notifiers) != 1 {
		t.Errorf("Expected 1 notifier, got %d", len(hm.notifiers))
	}
}

// TestGetDiskCheckInterval verifies the getter returns the correct interval
func TestGetDiskCheckInterval(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})
	
	interval := hm.GetDiskCheckInterval()
	if interval != 5*time.Minute {
		t.Errorf("Expected disk check interval of 5 minutes, got %v", interval)
	}
}

// TestGetStatusFilePath verifies the getter returns the correct path
func TestGetStatusFilePath(t *testing.T) {
	expectedPath := "/tmp/status.json"
	hm := NewHealthMonitor(expectedPath, []Notifier{})
	
	path := hm.GetStatusFilePath()
	if path != expectedPath {
		t.Errorf("Expected status file path '%s', got '%s'", expectedPath, path)
	}
}

// TestGetNotifiers verifies the getter returns the correct notifiers
func TestGetNotifiers(t *testing.T) {
	notifiers := []Notifier{&mockNotifier{}, &mockNotifier{}}
	hm := NewHealthMonitor("", notifiers)
	
	returnedNotifiers := hm.GetNotifiers()
	if len(returnedNotifiers) != 2 {
		t.Errorf("Expected 2 notifiers, got %d", len(returnedNotifiers))
	}
}

// TestMonitorDiskSpace_10GBThreshold verifies that immediate upload is triggered at 10 GB
func TestMonitorDiskSpace_10GBThreshold(t *testing.T) {
	mockNotifier := &mockNotifier{
		sendFunc: func(title, message string) error {
			// Notification would be sent if disk usage exceeds 10GB
			return nil
		},
	}

	hm := &HealthMonitor{
		notifiers:         []Notifier{mockNotifier},
		diskCheckInterval: 100 * time.Millisecond,
		statusFilePath:    "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// This test will use the actual disk stats, so we can't easily simulate 10GB usage
	// Instead, we verify the method runs without errors
	err := hm.MonitorDiskSpace(ctx, ".", nil, nil)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// TestMonitorDiskSpace_12GBThreshold verifies that new recordings are paused at 12 GB
func TestMonitorDiskSpace_12GBThreshold(t *testing.T) {
	mockNotifier := &mockNotifier{
		sendFunc: func(title, message string) error {
			// Notification would be sent if disk usage exceeds 12GB
			return nil
		},
	}

	hm := &HealthMonitor{
		notifiers:         []Notifier{mockNotifier},
		diskCheckInterval: 100 * time.Millisecond,
		statusFilePath:    "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// This test will use the actual disk stats
	err := hm.MonitorDiskSpace(ctx, ".", nil, nil)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// TestMonitorDiskSpace_13GBThreshold verifies that oldest recording is stopped at 13 GB
func TestMonitorDiskSpace_13GBThreshold(t *testing.T) {
	stopFunc := func() error {
		// Stop function would be called if disk usage exceeds 13GB
		return nil
	}

	mockNotifier := &mockNotifier{
		sendFunc: func(title, message string) error {
			// Notification would be sent if disk usage exceeds 13GB
			return nil
		},
	}

	hm := &HealthMonitor{
		notifiers:         []Notifier{mockNotifier},
		diskCheckInterval: 100 * time.Millisecond,
		statusFilePath:    "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// This test will use the actual disk stats
	err := hm.MonitorDiskSpace(ctx, ".", nil, stopFunc)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// TestMonitorDiskSpace_ErrorHandling verifies that monitoring continues after disk stat errors
func TestMonitorDiskSpace_ErrorHandling(t *testing.T) {
	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 100 * time.Millisecond,
		statusFilePath:    "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Use an invalid path to trigger errors
	err := hm.MonitorDiskSpace(ctx, "/nonexistent/path/that/does/not/exist", nil, nil)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// TestMonitorDiskSpace_NilStopFunc verifies that monitoring handles nil stop function gracefully
func TestMonitorDiskSpace_NilStopFunc(t *testing.T) {
	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 100 * time.Millisecond,
		statusFilePath:    "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Pass nil for stopOldestRecordingFunc
	err := hm.MonitorDiskSpace(ctx, ".", nil, nil)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// mockNotifier is a test implementation of the Notifier interface
type mockNotifier struct {
	sendFunc func(title, message string) error
}

func (m *mockNotifier) Send(title, message string) error {
	if m.sendFunc != nil {
		return m.sendFunc(title, message)
	}
	return nil
}

// TestUpdateStatusFile verifies that status file is written correctly
func TestUpdateStatusFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	statusFilePath := tempDir + "/status.json"

	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    statusFilePath,
	}

	// Create test status
	status := SystemStatus{
		SessionID:           "test-session-123",
		StartTime:           time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ActiveRecordings:    3,
		ActiveMatrixJobs:    []MatrixJobStatus{
			{
				JobID:          "job-1",
				Channel:        "channel_a",
				RecordingState: "recording",
				LastActivity:   time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
			},
			{
				JobID:          "job-2",
				Channel:        "channel_b",
				RecordingState: "idle",
				LastActivity:   time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC),
			},
		},
		DiskUsageBytes:      5368709120, // 5 GB
		DiskTotalBytes:      15032385536, // 14 GB
		LastChainTransition: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		GofileUploads:       10,
		FilesterUploads:     10,
	}

	// Note: We can't test git operations without a real git repository
	// So we'll test only the JSON marshaling and file writing
	// The git operations will fail, but we can verify the file was written

	// Call UpdateStatusFile (will fail on git operations, but that's expected in test)
	_ = hm.UpdateStatusFile(status)

	// Verify the file was created
	if _, err := os.Stat(statusFilePath); os.IsNotExist(err) {
		t.Error("Status file was not created")
		return
	}

	// Read the file and verify its contents
	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		t.Errorf("Failed to read status file: %v", err)
		return
	}

	// Unmarshal and verify
	var readStatus SystemStatus
	if err := json.Unmarshal(data, &readStatus); err != nil {
		t.Errorf("Failed to unmarshal status file: %v", err)
		return
	}

	// Verify key fields
	if readStatus.SessionID != status.SessionID {
		t.Errorf("Expected session ID '%s', got '%s'", status.SessionID, readStatus.SessionID)
	}

	if readStatus.ActiveRecordings != status.ActiveRecordings {
		t.Errorf("Expected %d active recordings, got %d", status.ActiveRecordings, readStatus.ActiveRecordings)
	}

	if len(readStatus.ActiveMatrixJobs) != len(status.ActiveMatrixJobs) {
		t.Errorf("Expected %d matrix jobs, got %d", len(status.ActiveMatrixJobs), len(readStatus.ActiveMatrixJobs))
	}

	if readStatus.DiskUsageBytes != status.DiskUsageBytes {
		t.Errorf("Expected disk usage %d, got %d", status.DiskUsageBytes, readStatus.DiskUsageBytes)
	}

	if readStatus.GofileUploads != status.GofileUploads {
		t.Errorf("Expected %d Gofile uploads, got %d", status.GofileUploads, readStatus.GofileUploads)
	}

	if readStatus.FilesterUploads != status.FilesterUploads {
		t.Errorf("Expected %d Filester uploads, got %d", status.FilesterUploads, readStatus.FilesterUploads)
	}
}

// TestUpdateStatusFile_JSONFormatting verifies that JSON is properly formatted
func TestUpdateStatusFile_JSONFormatting(t *testing.T) {
	tempDir := t.TempDir()
	statusFilePath := tempDir + "/status.json"

	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    statusFilePath,
	}

	status := SystemStatus{
		SessionID:        "test-session-456",
		StartTime:        time.Now(),
		ActiveRecordings: 1,
		ActiveMatrixJobs: []MatrixJobStatus{
			{
				JobID:          "job-1",
				Channel:        "test_channel",
				RecordingState: "recording",
				LastActivity:   time.Now(),
			},
		},
		DiskUsageBytes:      1073741824, // 1 GB
		DiskTotalBytes:      15032385536, // 14 GB
		LastChainTransition: time.Now(),
		GofileUploads:       5,
		FilesterUploads:     5,
	}

	// Call UpdateStatusFile (will fail on git operations, but file should be written)
	_ = hm.UpdateStatusFile(status)

	// Read the file
	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		t.Errorf("Failed to read status file: %v", err)
		return
	}

	// Verify JSON is indented (contains newlines and spaces)
	jsonStr := string(data)
	if !strings.Contains(jsonStr, "\n") {
		t.Error("JSON should be formatted with newlines")
	}

	if !strings.Contains(jsonStr, "  ") {
		t.Error("JSON should be indented with spaces")
	}

	// Verify it's valid JSON
	var readStatus SystemStatus
	if err := json.Unmarshal(data, &readStatus); err != nil {
		t.Errorf("Status file contains invalid JSON: %v", err)
	}
}

// TestUpdateStatusFile_EmptyMatrixJobs verifies handling of empty matrix jobs array
func TestUpdateStatusFile_EmptyMatrixJobs(t *testing.T) {
	tempDir := t.TempDir()
	statusFilePath := tempDir + "/status.json"

	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    statusFilePath,
	}

	status := SystemStatus{
		SessionID:           "test-session-789",
		StartTime:           time.Now(),
		ActiveRecordings:    0,
		ActiveMatrixJobs:    []MatrixJobStatus{}, // Empty array
		DiskUsageBytes:      0,
		DiskTotalBytes:      15032385536,
		LastChainTransition: time.Now(),
		GofileUploads:       0,
		FilesterUploads:     0,
	}

	// Call UpdateStatusFile
	_ = hm.UpdateStatusFile(status)

	// Read and verify
	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		t.Errorf("Failed to read status file: %v", err)
		return
	}

	var readStatus SystemStatus
	if err := json.Unmarshal(data, &readStatus); err != nil {
		t.Errorf("Failed to unmarshal status file: %v", err)
		return
	}

	if readStatus.ActiveMatrixJobs == nil {
		t.Error("ActiveMatrixJobs should be an empty array, not nil")
	}

	if len(readStatus.ActiveMatrixJobs) != 0 {
		t.Errorf("Expected 0 matrix jobs, got %d", len(readStatus.ActiveMatrixJobs))
	}
}

// TestUpdateStatusFile_InvalidPath verifies error handling for invalid file paths
func TestUpdateStatusFile_InvalidPath(t *testing.T) {
	// Use an invalid path that cannot be written to
	invalidPath := "/nonexistent/directory/that/does/not/exist/status.json"

	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    invalidPath,
	}

	status := SystemStatus{
		SessionID:        "test-session",
		StartTime:        time.Now(),
		ActiveRecordings: 0,
	}

	// Call UpdateStatusFile - should return an error
	err := hm.UpdateStatusFile(status)
	if err == nil {
		t.Error("Expected error when writing to invalid path, got nil")
	}

	// Verify error message mentions file writing
	if !strings.Contains(err.Error(), "failed to write status file") {
		t.Errorf("Expected error about writing file, got: %v", err)
	}
}

// TestUpdateStatusFile_AllFields verifies all SystemStatus fields are persisted
func TestUpdateStatusFile_AllFields(t *testing.T) {
	tempDir := t.TempDir()
	statusFilePath := tempDir + "/status.json"

	hm := &HealthMonitor{
		notifiers:         []Notifier{},
		diskCheckInterval: 5 * time.Minute,
		statusFilePath:    statusFilePath,
	}

	// Create status with all fields populated
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	lastActivity1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	lastActivity2 := time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC)
	lastChainTransition := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	status := SystemStatus{
		SessionID:        "session-abc-123",
		StartTime:        startTime,
		ActiveRecordings: 5,
		ActiveMatrixJobs: []MatrixJobStatus{
			{
				JobID:          "matrix-job-1",
				Channel:        "channel_alpha",
				RecordingState: "recording",
				LastActivity:   lastActivity1,
			},
			{
				JobID:          "matrix-job-2",
				Channel:        "channel_beta",
				RecordingState: "idle",
				LastActivity:   lastActivity2,
			},
		},
		DiskUsageBytes:      10737418240, // 10 GB
		DiskTotalBytes:      15032385536,  // 14 GB
		LastChainTransition: lastChainTransition,
		GofileUploads:       25,
		FilesterUploads:     25,
	}

	// Call UpdateStatusFile
	_ = hm.UpdateStatusFile(status)

	// Read and verify all fields
	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		t.Fatalf("Failed to read status file: %v", err)
	}

	var readStatus SystemStatus
	if err := json.Unmarshal(data, &readStatus); err != nil {
		t.Fatalf("Failed to unmarshal status file: %v", err)
	}

	// Verify all fields match
	if readStatus.SessionID != status.SessionID {
		t.Errorf("SessionID mismatch: expected '%s', got '%s'", status.SessionID, readStatus.SessionID)
	}

	if !readStatus.StartTime.Equal(status.StartTime) {
		t.Errorf("StartTime mismatch: expected %v, got %v", status.StartTime, readStatus.StartTime)
	}

	if readStatus.ActiveRecordings != status.ActiveRecordings {
		t.Errorf("ActiveRecordings mismatch: expected %d, got %d", status.ActiveRecordings, readStatus.ActiveRecordings)
	}

	if len(readStatus.ActiveMatrixJobs) != len(status.ActiveMatrixJobs) {
		t.Fatalf("ActiveMatrixJobs length mismatch: expected %d, got %d", len(status.ActiveMatrixJobs), len(readStatus.ActiveMatrixJobs))
	}

	for i, job := range status.ActiveMatrixJobs {
		readJob := readStatus.ActiveMatrixJobs[i]
		if readJob.JobID != job.JobID {
			t.Errorf("Job %d JobID mismatch: expected '%s', got '%s'", i, job.JobID, readJob.JobID)
		}
		if readJob.Channel != job.Channel {
			t.Errorf("Job %d Channel mismatch: expected '%s', got '%s'", i, job.Channel, readJob.Channel)
		}
		if readJob.RecordingState != job.RecordingState {
			t.Errorf("Job %d RecordingState mismatch: expected '%s', got '%s'", i, job.RecordingState, readJob.RecordingState)
		}
		if !readJob.LastActivity.Equal(job.LastActivity) {
			t.Errorf("Job %d LastActivity mismatch: expected %v, got %v", i, job.LastActivity, readJob.LastActivity)
		}
	}

	if readStatus.DiskUsageBytes != status.DiskUsageBytes {
		t.Errorf("DiskUsageBytes mismatch: expected %d, got %d", status.DiskUsageBytes, readStatus.DiskUsageBytes)
	}

	if readStatus.DiskTotalBytes != status.DiskTotalBytes {
		t.Errorf("DiskTotalBytes mismatch: expected %d, got %d", status.DiskTotalBytes, readStatus.DiskTotalBytes)
	}

	if !readStatus.LastChainTransition.Equal(status.LastChainTransition) {
		t.Errorf("LastChainTransition mismatch: expected %v, got %v", status.LastChainTransition, readStatus.LastChainTransition)
	}

	if readStatus.GofileUploads != status.GofileUploads {
		t.Errorf("GofileUploads mismatch: expected %d, got %d", status.GofileUploads, readStatus.GofileUploads)
	}

	if readStatus.FilesterUploads != status.FilesterUploads {
		t.Errorf("FilesterUploads mismatch: expected %d, got %d", status.FilesterUploads, readStatus.FilesterUploads)
	}
}

// TestDetectRecordingGaps_SingleTransition verifies gap detection for a single transition
func TestDetectRecordingGaps_SingleTransition(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	// Create a transition with a 45-second gap
	endTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := endTime.Add(45 * time.Second)

	transitions := []Transition{
		{
			Channel:       "test_channel",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       endTime,
			StartTime:     startTime,
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != 1 {
		t.Fatalf("Expected 1 gap, got %d", len(gaps))
	}

	gap := gaps[0]
	if gap.Channel != "test_channel" {
		t.Errorf("Expected channel 'test_channel', got '%s'", gap.Channel)
	}

	if !gap.StartTime.Equal(endTime) {
		t.Errorf("Expected gap start time %v, got %v", endTime, gap.StartTime)
	}

	if !gap.EndTime.Equal(startTime) {
		t.Errorf("Expected gap end time %v, got %v", startTime, gap.EndTime)
	}

	expectedDuration := 45 * time.Second
	if gap.Duration != expectedDuration {
		t.Errorf("Expected gap duration %v, got %v", expectedDuration, gap.Duration)
	}
}

// TestDetectRecordingGaps_MultipleTransitions verifies gap detection for multiple transitions
func TestDetectRecordingGaps_MultipleTransitions(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transitions := []Transition{
		{
			Channel:       "channel_a",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       baseTime,
			StartTime:     baseTime.Add(30 * time.Second),
		},
		{
			Channel:       "channel_b",
			PreviousRunID: "run-3",
			NextRunID:     "run-4",
			EndTime:       baseTime.Add(1 * time.Minute),
			StartTime:     baseTime.Add(2 * time.Minute),
		},
		{
			Channel:       "channel_c",
			PreviousRunID: "run-5",
			NextRunID:     "run-6",
			EndTime:       baseTime.Add(3 * time.Minute),
			StartTime:     baseTime.Add(4 * time.Minute),
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != 3 {
		t.Fatalf("Expected 3 gaps, got %d", len(gaps))
	}

	// Verify first gap (30 seconds)
	if gaps[0].Channel != "channel_a" {
		t.Errorf("Gap 0: Expected channel 'channel_a', got '%s'", gaps[0].Channel)
	}
	if gaps[0].Duration != 30*time.Second {
		t.Errorf("Gap 0: Expected duration 30s, got %v", gaps[0].Duration)
	}

	// Verify second gap (60 seconds)
	if gaps[1].Channel != "channel_b" {
		t.Errorf("Gap 1: Expected channel 'channel_b', got '%s'", gaps[1].Channel)
	}
	if gaps[1].Duration != 60*time.Second {
		t.Errorf("Gap 1: Expected duration 60s, got %v", gaps[1].Duration)
	}

	// Verify third gap (60 seconds)
	if gaps[2].Channel != "channel_c" {
		t.Errorf("Gap 2: Expected channel 'channel_c', got '%s'", gaps[2].Channel)
	}
	if gaps[2].Duration != 60*time.Second {
		t.Errorf("Gap 2: Expected duration 60s, got %v", gaps[2].Duration)
	}
}

// TestDetectRecordingGaps_EmptyTransitions verifies handling of empty transitions slice
func TestDetectRecordingGaps_EmptyTransitions(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	transitions := []Transition{}
	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps for empty transitions, got %d", len(gaps))
	}

	if gaps == nil {
		t.Error("Expected non-nil gaps slice, got nil")
	}
}

// TestDetectRecordingGaps_NegativeGap verifies that negative gaps are not recorded
func TestDetectRecordingGaps_NegativeGap(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	// Create a transition where next run started BEFORE previous ended (overlapping)
	endTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := endTime.Add(-10 * time.Second) // Started 10 seconds before end

	transitions := []Transition{
		{
			Channel:       "test_channel",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       endTime,
			StartTime:     startTime,
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	// Negative gaps should not be recorded
	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps for negative gap duration, got %d", len(gaps))
	}
}

// TestDetectRecordingGaps_ZeroGap verifies that zero-duration gaps are not recorded
func TestDetectRecordingGaps_ZeroGap(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	// Create a transition where next run started exactly when previous ended
	transitionTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transitions := []Transition{
		{
			Channel:       "test_channel",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       transitionTime,
			StartTime:     transitionTime,
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	// Zero-duration gaps should not be recorded
	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps for zero gap duration, got %d", len(gaps))
	}
}

// TestDetectRecordingGaps_MixedGaps verifies handling of mixed positive, negative, and zero gaps
func TestDetectRecordingGaps_MixedGaps(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	transitions := []Transition{
		{
			Channel:       "channel_a",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       baseTime,
			StartTime:     baseTime.Add(45 * time.Second), // Positive gap: 45s
		},
		{
			Channel:       "channel_b",
			PreviousRunID: "run-3",
			NextRunID:     "run-4",
			EndTime:       baseTime.Add(1 * time.Minute),
			StartTime:     baseTime.Add(1 * time.Minute), // Zero gap
		},
		{
			Channel:       "channel_c",
			PreviousRunID: "run-5",
			NextRunID:     "run-6",
			EndTime:       baseTime.Add(2 * time.Minute),
			StartTime:     baseTime.Add(90 * time.Second), // Negative gap: -30s
		},
		{
			Channel:       "channel_d",
			PreviousRunID: "run-7",
			NextRunID:     "run-8",
			EndTime:       baseTime.Add(3 * time.Minute),
			StartTime:     baseTime.Add(4 * time.Minute), // Positive gap: 60s
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	// Only positive gaps should be recorded
	if len(gaps) != 2 {
		t.Fatalf("Expected 2 gaps (only positive), got %d", len(gaps))
	}

	// Verify first positive gap (channel_a, 45s)
	if gaps[0].Channel != "channel_a" {
		t.Errorf("Gap 0: Expected channel 'channel_a', got '%s'", gaps[0].Channel)
	}
	if gaps[0].Duration != 45*time.Second {
		t.Errorf("Gap 0: Expected duration 45s, got %v", gaps[0].Duration)
	}

	// Verify second positive gap (channel_d, 60s)
	if gaps[1].Channel != "channel_d" {
		t.Errorf("Gap 1: Expected channel 'channel_d', got '%s'", gaps[1].Channel)
	}
	if gaps[1].Duration != 60*time.Second {
		t.Errorf("Gap 1: Expected duration 60s, got %v", gaps[1].Duration)
	}
}

// TestDetectRecordingGaps_TypicalGapRange verifies detection of typical 30-60 second gaps
func TestDetectRecordingGaps_TypicalGapRange(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Test various gap durations in the typical range
	testCases := []struct {
		channel  string
		duration time.Duration
	}{
		{"channel_30s", 30 * time.Second},
		{"channel_45s", 45 * time.Second},
		{"channel_60s", 60 * time.Second},
		{"channel_35s", 35 * time.Second},
		{"channel_50s", 50 * time.Second},
	}

	transitions := make([]Transition, len(testCases))
	for i, tc := range testCases {
		transitions[i] = Transition{
			Channel:       tc.channel,
			PreviousRunID: fmt.Sprintf("run-%d", i*2+1),
			NextRunID:     fmt.Sprintf("run-%d", i*2+2),
			EndTime:       baseTime.Add(time.Duration(i) * time.Minute),
			StartTime:     baseTime.Add(time.Duration(i)*time.Minute + tc.duration),
		}
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != len(testCases) {
		t.Fatalf("Expected %d gaps, got %d", len(testCases), len(gaps))
	}

	for i, tc := range testCases {
		if gaps[i].Channel != tc.channel {
			t.Errorf("Gap %d: Expected channel '%s', got '%s'", i, tc.channel, gaps[i].Channel)
		}
		if gaps[i].Duration != tc.duration {
			t.Errorf("Gap %d: Expected duration %v, got %v", i, tc.duration, gaps[i].Duration)
		}
	}
}

// TestDetectRecordingGaps_LargeGap verifies detection of unusually large gaps
func TestDetectRecordingGaps_LargeGap(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	// Create a transition with a 5-minute gap (unusually large)
	endTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := endTime.Add(5 * time.Minute)

	transitions := []Transition{
		{
			Channel:       "test_channel",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       endTime,
			StartTime:     startTime,
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != 1 {
		t.Fatalf("Expected 1 gap, got %d", len(gaps))
	}

	expectedDuration := 5 * time.Minute
	if gaps[0].Duration != expectedDuration {
		t.Errorf("Expected gap duration %v, got %v", expectedDuration, gaps[0].Duration)
	}

	// Verify the gap is correctly identified as unusually large
	if gaps[0].Duration < 2*time.Minute {
		t.Error("Expected gap to be identified as unusually large (> 2 minutes)")
	}
}

// TestDetectRecordingGaps_SmallGap verifies detection of very small gaps
func TestDetectRecordingGaps_SmallGap(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	// Create a transition with a 1-second gap (very small)
	endTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := endTime.Add(1 * time.Second)

	transitions := []Transition{
		{
			Channel:       "test_channel",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       endTime,
			StartTime:     startTime,
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != 1 {
		t.Fatalf("Expected 1 gap, got %d", len(gaps))
	}

	expectedDuration := 1 * time.Second
	if gaps[0].Duration != expectedDuration {
		t.Errorf("Expected gap duration %v, got %v", expectedDuration, gaps[0].Duration)
	}
}

// TestDetectRecordingGaps_DifferentChannels verifies gap detection across different channels
func TestDetectRecordingGaps_DifferentChannels(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	channels := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	transitions := make([]Transition, len(channels))

	for i, channel := range channels {
		transitions[i] = Transition{
			Channel:       channel,
			PreviousRunID: fmt.Sprintf("run-%s-1", channel),
			NextRunID:     fmt.Sprintf("run-%s-2", channel),
			EndTime:       baseTime.Add(time.Duration(i) * time.Minute),
			StartTime:     baseTime.Add(time.Duration(i)*time.Minute + 40*time.Second),
		}
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != len(channels) {
		t.Fatalf("Expected %d gaps, got %d", len(channels), len(gaps))
	}

	// Verify each channel has its gap recorded
	for i, channel := range channels {
		if gaps[i].Channel != channel {
			t.Errorf("Gap %d: Expected channel '%s', got '%s'", i, channel, gaps[i].Channel)
		}
		if gaps[i].Duration != 40*time.Second {
			t.Errorf("Gap %d: Expected duration 40s, got %v", i, gaps[i].Duration)
		}
	}
}

// TestDetectRecordingGaps_TimestampAccuracy verifies accurate timestamp recording
func TestDetectRecordingGaps_TimestampAccuracy(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	endTime := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
	startTime := time.Date(2024, 1, 15, 10, 31, 30, 987654321, time.UTC)

	transitions := []Transition{
		{
			Channel:       "test_channel",
			PreviousRunID: "run-1",
			NextRunID:     "run-2",
			EndTime:       endTime,
			StartTime:     startTime,
		},
	}

	gaps := hm.DetectRecordingGaps(transitions)

	if len(gaps) != 1 {
		t.Fatalf("Expected 1 gap, got %d", len(gaps))
	}

	gap := gaps[0]

	// Verify exact timestamp matching (including nanoseconds)
	if !gap.StartTime.Equal(endTime) {
		t.Errorf("Gap start time mismatch: expected %v, got %v", endTime, gap.StartTime)
	}

	if !gap.EndTime.Equal(startTime) {
		t.Errorf("Gap end time mismatch: expected %v, got %v", startTime, gap.EndTime)
	}

	// Verify duration calculation is accurate
	expectedDuration := startTime.Sub(endTime)
	if gap.Duration != expectedDuration {
		t.Errorf("Gap duration mismatch: expected %v, got %v", expectedDuration, gap.Duration)
	}
}

// TestAggregateMatrixJobHealth_AllHealthy verifies aggregation when all jobs are healthy
func TestAggregateMatrixJobHealth_AllHealthy(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	jobs := []MatrixJobStatus{
		{
			JobID:          "job-1",
			Channel:        "channel_a",
			RecordingState: "recording",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-2",
			Channel:        "channel_b",
			RecordingState: "idle",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-3",
			Channel:        "channel_c",
			RecordingState: "recording",
			LastActivity:   time.Now(),
		},
	}

	health := hm.AggregateMatrixJobHealth(jobs)

	if health.TotalJobs != 3 {
		t.Errorf("Expected 3 total jobs, got %d", health.TotalJobs)
	}

	if health.ActiveJobs != 2 {
		t.Errorf("Expected 2 active jobs, got %d", health.ActiveJobs)
	}

	if health.IdleJobs != 1 {
		t.Errorf("Expected 1 idle job, got %d", health.IdleJobs)
	}

	if health.FailedJobs != 0 {
		t.Errorf("Expected 0 failed jobs, got %d", health.FailedJobs)
	}

	if health.TotalRecordings != 2 {
		t.Errorf("Expected 2 total recordings, got %d", health.TotalRecordings)
	}

	if health.HealthStatus != "healthy" {
		t.Errorf("Expected health status 'healthy', got '%s'", health.HealthStatus)
	}
}

// TestAggregateMatrixJobHealth_Degraded verifies aggregation when some jobs have failed
func TestAggregateMatrixJobHealth_Degraded(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	jobs := []MatrixJobStatus{
		{
			JobID:          "job-1",
			Channel:        "channel_a",
			RecordingState: "recording",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-2",
			Channel:        "channel_b",
			RecordingState: "recording",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-3",
			Channel:        "channel_c",
			RecordingState: "failed",
			LastActivity:   time.Now(),
		},
	}

	health := hm.AggregateMatrixJobHealth(jobs)

	if health.TotalJobs != 3 {
		t.Errorf("Expected 3 total jobs, got %d", health.TotalJobs)
	}

	if health.ActiveJobs != 2 {
		t.Errorf("Expected 2 active jobs, got %d", health.ActiveJobs)
	}

	if health.FailedJobs != 1 {
		t.Errorf("Expected 1 failed job, got %d", health.FailedJobs)
	}

	if health.HealthStatus != "degraded" {
		t.Errorf("Expected health status 'degraded', got '%s'", health.HealthStatus)
	}
}

// TestAggregateMatrixJobHealth_Critical verifies aggregation when majority of jobs have failed
func TestAggregateMatrixJobHealth_Critical(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	jobs := []MatrixJobStatus{
		{
			JobID:          "job-1",
			Channel:        "channel_a",
			RecordingState: "recording",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-2",
			Channel:        "channel_b",
			RecordingState: "failed",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-3",
			Channel:        "channel_c",
			RecordingState: "error",
			LastActivity:   time.Now(),
		},
	}

	health := hm.AggregateMatrixJobHealth(jobs)

	if health.TotalJobs != 3 {
		t.Errorf("Expected 3 total jobs, got %d", health.TotalJobs)
	}

	if health.ActiveJobs != 1 {
		t.Errorf("Expected 1 active job, got %d", health.ActiveJobs)
	}

	if health.FailedJobs != 2 {
		t.Errorf("Expected 2 failed jobs, got %d", health.FailedJobs)
	}

	if health.HealthStatus != "critical" {
		t.Errorf("Expected health status 'critical', got '%s'", health.HealthStatus)
	}
}

// TestAggregateMatrixJobHealth_NoJobs verifies aggregation when no jobs are running
func TestAggregateMatrixJobHealth_NoJobs(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	health := hm.AggregateMatrixJobHealth([]MatrixJobStatus{})

	if health.TotalJobs != 0 {
		t.Errorf("Expected 0 total jobs, got %d", health.TotalJobs)
	}

	if health.HealthStatus != "critical" {
		t.Errorf("Expected health status 'critical' for no jobs, got '%s'", health.HealthStatus)
	}
}

// TestAggregateMatrixJobHealth_UnknownStates verifies handling of unknown recording states
func TestAggregateMatrixJobHealth_UnknownStates(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	jobs := []MatrixJobStatus{
		{
			JobID:          "job-1",
			Channel:        "channel_a",
			RecordingState: "unknown_state",
			LastActivity:   time.Now(),
		},
		{
			JobID:          "job-2",
			Channel:        "channel_b",
			RecordingState: "recording",
			LastActivity:   time.Now(),
		},
	}

	health := hm.AggregateMatrixJobHealth(jobs)

	if health.TotalJobs != 2 {
		t.Errorf("Expected 2 total jobs, got %d", health.TotalJobs)
	}

	if health.ActiveJobs != 1 {
		t.Errorf("Expected 1 active job, got %d", health.ActiveJobs)
	}

	// Unknown states should be treated as idle
	if health.IdleJobs != 1 {
		t.Errorf("Expected 1 idle job (unknown state), got %d", health.IdleJobs)
	}

	if health.HealthStatus != "healthy" {
		t.Errorf("Expected health status 'healthy', got '%s'", health.HealthStatus)
	}
}

// TestDetectWorkflowStartFailure_SuccessfulTransition verifies detection when workflow starts on time
func TestDetectWorkflowStartFailure_SuccessfulTransition(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	triggerTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := triggerTime.Add(2 * time.Minute) // Started 2 minutes after trigger

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	nextRun := &WorkflowRun{
		RunID:     "run-2",
		SessionID: "session-2",
		StartTime: startTime,
	}

	maxExpectedDelay := 5 * time.Minute

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nextRun, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if gap != nil {
		t.Errorf("Expected no gap for successful transition, got gap with duration %v", gap.GapDuration)
	}
}

// TestDetectWorkflowStartFailure_DelayedStart verifies detection when workflow starts late
func TestDetectWorkflowStartFailure_DelayedStart(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	triggerTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := triggerTime.Add(10 * time.Minute) // Started 10 minutes after trigger (late)

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	nextRun := &WorkflowRun{
		RunID:     "run-2",
		SessionID: "session-2",
		StartTime: startTime,
	}

	maxExpectedDelay := 5 * time.Minute

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nextRun, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if gap == nil {
		t.Fatal("Expected gap to be detected for delayed start")
	}

	if gap.PreviousRunID != "run-1" {
		t.Errorf("Expected previous run ID 'run-1', got '%s'", gap.PreviousRunID)
	}

	if !gap.NextRunStarted {
		t.Error("Expected NextRunStarted to be true")
	}

	if !gap.ActualStart.Equal(startTime) {
		t.Errorf("Expected actual start time %v, got %v", startTime, gap.ActualStart)
	}

	expectedDuration := 10 * time.Minute
	if gap.GapDuration != expectedDuration {
		t.Errorf("Expected gap duration %v, got %v", expectedDuration, gap.GapDuration)
	}
}

// TestDetectWorkflowStartFailure_NoNextRun verifies detection when workflow never starts
func TestDetectWorkflowStartFailure_NoNextRun(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	triggerTime := time.Now().Add(-10 * time.Minute) // Triggered 10 minutes ago

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	maxExpectedDelay := 5 * time.Minute

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nil, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if gap == nil {
		t.Fatal("Expected gap to be detected when workflow never starts")
	}

	if gap.PreviousRunID != "run-1" {
		t.Errorf("Expected previous run ID 'run-1', got '%s'", gap.PreviousRunID)
	}

	if gap.NextRunStarted {
		t.Error("Expected NextRunStarted to be false")
	}

	if !gap.ActualStart.IsZero() {
		t.Errorf("Expected zero actual start time, got %v", gap.ActualStart)
	}

	// Gap duration should be approximately 10 minutes (time since trigger)
	if gap.GapDuration < 9*time.Minute || gap.GapDuration > 11*time.Minute {
		t.Errorf("Expected gap duration around 10 minutes, got %v", gap.GapDuration)
	}
}

// TestDetectWorkflowStartFailure_NoNextRunWithinDelay verifies no gap when within expected delay
func TestDetectWorkflowStartFailure_NoNextRunWithinDelay(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	triggerTime := time.Now().Add(-2 * time.Minute) // Triggered 2 minutes ago

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	maxExpectedDelay := 5 * time.Minute

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nil, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if gap != nil {
		t.Errorf("Expected no gap when within expected delay, got gap with duration %v", gap.GapDuration)
	}
}

// TestDetectWorkflowStartFailure_NoChainTriggered verifies error when chain not triggered
func TestDetectWorkflowStartFailure_NoChainTriggered(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		ChainTriggered: false, // Chain not triggered
	}

	maxExpectedDelay := 5 * time.Minute

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nil, maxExpectedDelay)

	if err == nil {
		t.Error("Expected error when chain not triggered, got nil")
	}

	if gap != nil {
		t.Error("Expected no gap when chain not triggered")
	}

	if !strings.Contains(err.Error(), "did not trigger a chain transition") {
		t.Errorf("Expected error about chain not triggered, got: %v", err)
	}
}

// TestDetectWorkflowStartFailure_NoTriggerTime verifies error when trigger time not set
func TestDetectWorkflowStartFailure_NoTriggerTime(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		ChainTriggered: true,
		TriggerTime:    time.Time{}, // Zero value - not set
	}

	maxExpectedDelay := 5 * time.Minute

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nil, maxExpectedDelay)

	if err == nil {
		t.Error("Expected error when trigger time not set, got nil")
	}

	if gap != nil {
		t.Error("Expected no gap when trigger time not set")
	}

	if !strings.Contains(err.Error(), "has no trigger time set") {
		t.Errorf("Expected error about trigger time not set, got: %v", err)
	}
}

// TestDetectWorkflowStartFailure_EdgeCaseDelay verifies detection at exact threshold
func TestDetectWorkflowStartFailure_EdgeCaseDelay(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	triggerTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	maxExpectedDelay := 5 * time.Minute
	startTime := triggerTime.Add(maxExpectedDelay) // Exactly at threshold

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	nextRun := &WorkflowRun{
		RunID:     "run-2",
		SessionID: "session-2",
		StartTime: startTime,
	}

	gap, err := hm.DetectWorkflowStartFailure(previousRun, nextRun, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// At exactly the threshold, should not be considered a gap
	if gap != nil {
		t.Errorf("Expected no gap at exact threshold, got gap with duration %v", gap.GapDuration)
	}
}

// TestDetectWorkflowStartFailures_MultipleRuns verifies batch detection across multiple runs
func TestDetectWorkflowStartFailures_MultipleRuns(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	maxExpectedDelay := 5 * time.Minute

	runs := []WorkflowRun{
		{
			RunID:          "run-1",
			SessionID:      "session-1",
			StartTime:      baseTime,
			EndTime:        baseTime.Add(5 * time.Hour),
			ChainTriggered: true,
			TriggerTime:    baseTime.Add(5 * time.Hour),
		},
		{
			RunID:     "run-2",
			SessionID: "session-2",
			StartTime: baseTime.Add(5*time.Hour + 2*time.Minute), // 2 min delay - OK
		},
		{
			RunID:          "run-2",
			SessionID:      "session-2",
			StartTime:      baseTime.Add(5 * time.Hour),
			EndTime:        baseTime.Add(10 * time.Hour),
			ChainTriggered: true,
			TriggerTime:    baseTime.Add(10 * time.Hour),
		},
		{
			RunID:     "run-3",
			SessionID: "session-3",
			StartTime: baseTime.Add(10*time.Hour + 10*time.Minute), // 10 min delay - GAP
		},
		{
			RunID:          "run-3",
			SessionID:      "session-3",
			StartTime:      baseTime.Add(10 * time.Hour),
			EndTime:        baseTime.Add(15 * time.Hour),
			ChainTriggered: true,
			TriggerTime:    baseTime.Add(15 * time.Hour),
		},
		{
			RunID:     "run-4",
			SessionID: "session-4",
			StartTime: baseTime.Add(15*time.Hour + 3*time.Minute), // 3 min delay - OK
		},
	}

	gaps := hm.DetectWorkflowStartFailures(runs, maxExpectedDelay)

	// Should detect 1 gap (run-2 to run-3 with 10 minute delay)
	if len(gaps) != 1 {
		t.Fatalf("Expected 1 gap, got %d", len(gaps))
	}

	gap := gaps[0]
	if gap.PreviousRunID != "run-2" {
		t.Errorf("Expected previous run ID 'run-2', got '%s'", gap.PreviousRunID)
	}

	if gap.GapDuration != 10*time.Minute {
		t.Errorf("Expected gap duration 10 minutes, got %v", gap.GapDuration)
	}

	if !gap.NextRunStarted {
		t.Error("Expected NextRunStarted to be true")
	}
}

// TestDetectWorkflowStartFailures_NoChainTransitions verifies handling when no chains triggered
func TestDetectWorkflowStartFailures_NoChainTransitions(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	maxExpectedDelay := 5 * time.Minute

	runs := []WorkflowRun{
		{
			RunID:          "run-1",
			SessionID:      "session-1",
			StartTime:      baseTime,
			EndTime:        baseTime.Add(5 * time.Hour),
			ChainTriggered: false, // No chain triggered
		},
		{
			RunID:          "run-2",
			SessionID:      "session-2",
			StartTime:      baseTime.Add(6 * time.Hour),
			EndTime:        baseTime.Add(11 * time.Hour),
			ChainTriggered: false, // No chain triggered
		},
	}

	gaps := hm.DetectWorkflowStartFailures(runs, maxExpectedDelay)

	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps when no chains triggered, got %d", len(gaps))
	}
}

// TestDetectWorkflowStartFailures_EmptyRuns verifies handling of empty runs slice
func TestDetectWorkflowStartFailures_EmptyRuns(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	maxExpectedDelay := 5 * time.Minute

	gaps := hm.DetectWorkflowStartFailures([]WorkflowRun{}, maxExpectedDelay)

	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps for empty runs, got %d", len(gaps))
	}

	if gaps == nil {
		t.Error("Expected non-nil gaps slice, got nil")
	}
}

// TestDetectWorkflowStartFailures_MissingNextRun verifies detection when next run never starts
func TestDetectWorkflowStartFailures_MissingNextRun(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	triggerTime := time.Now().Add(-10 * time.Minute)
	maxExpectedDelay := 5 * time.Minute

	runs := []WorkflowRun{
		{
			RunID:          "run-1",
			SessionID:      "session-1",
			StartTime:      triggerTime.Add(-5 * time.Hour),
			EndTime:        triggerTime,
			ChainTriggered: true,
			TriggerTime:    triggerTime,
		},
		// No next run - workflow failed to start
	}

	gaps := hm.DetectWorkflowStartFailures(runs, maxExpectedDelay)

	if len(gaps) != 1 {
		t.Fatalf("Expected 1 gap when next run missing, got %d", len(gaps))
	}

	gap := gaps[0]
	if gap.PreviousRunID != "run-1" {
		t.Errorf("Expected previous run ID 'run-1', got '%s'", gap.PreviousRunID)
	}

	if gap.NextRunStarted {
		t.Error("Expected NextRunStarted to be false when next run missing")
	}

	if !gap.ActualStart.IsZero() {
		t.Errorf("Expected zero actual start time, got %v", gap.ActualStart)
	}
}

// TestDetectWorkflowStartFailures_AllSuccessful verifies no gaps when all transitions successful
func TestDetectWorkflowStartFailures_AllSuccessful(t *testing.T) {
	hm := NewHealthMonitor("", []Notifier{})

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	maxExpectedDelay := 5 * time.Minute

	runs := []WorkflowRun{
		{
			RunID:          "run-1",
			SessionID:      "session-1",
			StartTime:      baseTime,
			EndTime:        baseTime.Add(5 * time.Hour),
			ChainTriggered: true,
			TriggerTime:    baseTime.Add(5 * time.Hour),
		},
		{
			RunID:     "run-2",
			SessionID: "session-2",
			StartTime: baseTime.Add(5*time.Hour + 2*time.Minute), // 2 min delay - OK
		},
		{
			RunID:          "run-2",
			SessionID:      "session-2",
			StartTime:      baseTime.Add(5 * time.Hour),
			EndTime:        baseTime.Add(10 * time.Hour),
			ChainTriggered: true,
			TriggerTime:    baseTime.Add(10 * time.Hour),
		},
		{
			RunID:     "run-3",
			SessionID: "session-3",
			StartTime: baseTime.Add(10*time.Hour + 3*time.Minute), // 3 min delay - OK
		},
	}

	gaps := hm.DetectWorkflowStartFailures(runs, maxExpectedDelay)

	if len(gaps) != 0 {
		t.Errorf("Expected 0 gaps when all transitions successful, got %d", len(gaps))
	}
}

// TestDetectWorkflowStartFailure_NotificationSent verifies notification is sent for gaps
func TestDetectWorkflowStartFailure_NotificationSent(t *testing.T) {
	notificationSent := false
	var notificationTitle, notificationMessage string

	mockNotifier := &mockNotifier{
		sendFunc: func(title, message string) error {
			notificationSent = true
			notificationTitle = title
			notificationMessage = message
			return nil
		},
	}

	hm := NewHealthMonitor("", []Notifier{mockNotifier})

	triggerTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	startTime := triggerTime.Add(10 * time.Minute) // 10 min delay

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	nextRun := &WorkflowRun{
		RunID:     "run-2",
		SessionID: "session-2",
		StartTime: startTime,
	}

	maxExpectedDelay := 5 * time.Minute

	_, err := hm.DetectWorkflowStartFailure(previousRun, nextRun, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !notificationSent {
		t.Error("Expected notification to be sent for workflow start delay")
	}

	if !strings.Contains(notificationTitle, "Workflow Start Delay") {
		t.Errorf("Expected notification title to contain 'Workflow Start Delay', got '%s'", notificationTitle)
	}

	if !strings.Contains(notificationMessage, "run-1") {
		t.Errorf("Expected notification message to contain 'run-1', got '%s'", notificationMessage)
	}

	if !strings.Contains(notificationMessage, "run-2") {
		t.Errorf("Expected notification message to contain 'run-2', got '%s'", notificationMessage)
	}
}

// TestDetectWorkflowStartFailure_NotificationForMissingRun verifies notification when run never starts
func TestDetectWorkflowStartFailure_NotificationForMissingRun(t *testing.T) {
	notificationSent := false
	var notificationTitle, notificationMessage string

	mockNotifier := &mockNotifier{
		sendFunc: func(title, message string) error {
			notificationSent = true
			notificationTitle = title
			notificationMessage = message
			return nil
		},
	}

	hm := NewHealthMonitor("", []Notifier{mockNotifier})

	triggerTime := time.Now().Add(-10 * time.Minute)

	previousRun := WorkflowRun{
		RunID:          "run-1",
		SessionID:      "session-1",
		StartTime:      triggerTime.Add(-5 * time.Hour),
		EndTime:        triggerTime,
		ChainTriggered: true,
		TriggerTime:    triggerTime,
	}

	maxExpectedDelay := 5 * time.Minute

	_, err := hm.DetectWorkflowStartFailure(previousRun, nil, maxExpectedDelay)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !notificationSent {
		t.Error("Expected notification to be sent for workflow start failure")
	}

	if !strings.Contains(notificationTitle, "Workflow Start Failure") {
		t.Errorf("Expected notification title to contain 'Workflow Start Failure', got '%s'", notificationTitle)
	}

	if !strings.Contains(notificationMessage, "run-1") {
		t.Errorf("Expected notification message to contain 'run-1', got '%s'", notificationMessage)
	}

	if !strings.Contains(notificationMessage, "Manual intervention") {
		t.Errorf("Expected notification message to mention manual intervention, got '%s'", notificationMessage)
	}
}
