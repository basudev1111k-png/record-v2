package github_actions

import (
	"fmt"
	"testing"
	"time"
)

// TestNewMatrixCoordinator verifies that NewMatrixCoordinator creates a valid instance
func TestNewMatrixCoordinator(t *testing.T) {
	sessionID := "test-session-123"
	mc := NewMatrixCoordinator(sessionID)

	if mc == nil {
		t.Fatal("NewMatrixCoordinator returned nil")
	}

	if mc.sessionID != sessionID {
		t.Errorf("Expected sessionID %s, got %s", sessionID, mc.sessionID)
	}

	if mc.jobRegistry == nil {
		t.Error("jobRegistry should be initialized")
	}

	if len(mc.jobRegistry) != 0 {
		t.Errorf("Expected empty jobRegistry, got %d entries", len(mc.jobRegistry))
	}
}

// TestGetSessionID verifies that GetSessionID returns the correct session identifier
func TestGetSessionID(t *testing.T) {
	sessionID := "test-session-456"
	mc := NewMatrixCoordinator(sessionID)

	if mc.GetSessionID() != sessionID {
		t.Errorf("Expected sessionID %s, got %s", sessionID, mc.GetSessionID())
	}
}

// TestAssignChannels_ValidInput verifies that AssignChannels correctly distributes channels
func TestAssignChannels_ValidInput(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	tests := []struct {
		name           string
		channels       []string
		maxJobs        int
		expectedCount  int
		expectedJobIDs []string
	}{
		{
			name:           "single channel",
			channels:       []string{"channel1"},
			maxJobs:        5,
			expectedCount:  1,
			expectedJobIDs: []string{"matrix-job-1"},
		},
		{
			name:           "three channels",
			channels:       []string{"channel1", "channel2", "channel3"},
			maxJobs:        5,
			expectedCount:  3,
			expectedJobIDs: []string{"matrix-job-1", "matrix-job-2", "matrix-job-3"},
		},
		{
			name:           "five channels",
			channels:       []string{"ch1", "ch2", "ch3", "ch4", "ch5"},
			maxJobs:        10,
			expectedCount:  5,
			expectedJobIDs: []string{"matrix-job-1", "matrix-job-2", "matrix-job-3", "matrix-job-4", "matrix-job-5"},
		},
		{
			name:           "maximum 20 channels",
			channels:       []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10", "c11", "c12", "c13", "c14", "c15", "c16", "c17", "c18", "c19", "c20"},
			maxJobs:        20,
			expectedCount:  20,
			expectedJobIDs: []string{"matrix-job-1", "matrix-job-2", "matrix-job-3", "matrix-job-4", "matrix-job-5", "matrix-job-6", "matrix-job-7", "matrix-job-8", "matrix-job-9", "matrix-job-10", "matrix-job-11", "matrix-job-12", "matrix-job-13", "matrix-job-14", "matrix-job-15", "matrix-job-16", "matrix-job-17", "matrix-job-18", "matrix-job-19", "matrix-job-20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignments, err := mc.AssignChannels(tt.channels, tt.maxJobs)

			if err != nil {
				t.Fatalf("AssignChannels returned unexpected error: %v", err)
			}

			if len(assignments) != tt.expectedCount {
				t.Errorf("Expected %d assignments, got %d", tt.expectedCount, len(assignments))
			}

			// Verify each assignment has correct structure
			for i, assignment := range assignments {
				if assignment.JobID != tt.expectedJobIDs[i] {
					t.Errorf("Assignment %d: expected JobID %s, got %s", i, tt.expectedJobIDs[i], assignment.JobID)
				}

				if assignment.Channel != tt.channels[i] {
					t.Errorf("Assignment %d: expected Channel %s, got %s", i, tt.channels[i], assignment.Channel)
				}
			}

			// Verify exactly one channel per job
			jobIDMap := make(map[string]bool)
			channelMap := make(map[string]bool)
			for _, assignment := range assignments {
				if jobIDMap[assignment.JobID] {
					t.Errorf("Duplicate JobID found: %s", assignment.JobID)
				}
				jobIDMap[assignment.JobID] = true

				if channelMap[assignment.Channel] {
					t.Errorf("Duplicate Channel found: %s", assignment.Channel)
				}
				channelMap[assignment.Channel] = true
			}
		})
	}
}

// TestAssignChannels_ExceedsChannelLimit verifies validation when channel count exceeds 20
func TestAssignChannels_ExceedsChannelLimit(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Create 21 channels (exceeds limit)
	channels := make([]string, 21)
	for i := 0; i < 21; i++ {
		channels[i] = "channel" + string(rune('A'+i))
	}

	assignments, err := mc.AssignChannels(channels, 25)

	if err == nil {
		t.Fatal("Expected error for channel count exceeding 20, got nil")
	}

	if assignments != nil {
		t.Errorf("Expected nil assignments on error, got %d assignments", len(assignments))
	}

	expectedError := "channel count (21) exceeds GitHub Actions limit of 20"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestAssignChannels_InvalidMaxJobs verifies validation of maxJobs parameter
func TestAssignChannels_InvalidMaxJobs(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")
	channels := []string{"channel1", "channel2"}

	tests := []struct {
		name          string
		maxJobs       int
		expectedError string
	}{
		{
			name:          "maxJobs is zero",
			maxJobs:       0,
			expectedError: "maxJobs must be at least 1, got 0",
		},
		{
			name:          "maxJobs is negative",
			maxJobs:       -5,
			expectedError: "maxJobs must be at least 1, got -5",
		},
		{
			name:          "maxJobs exceeds 20",
			maxJobs:       21,
			expectedError: "maxJobs (21) cannot exceed GitHub Actions limit of 20",
		},
		{
			name:          "maxJobs is 25",
			maxJobs:       25,
			expectedError: "maxJobs (25) cannot exceed GitHub Actions limit of 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignments, err := mc.AssignChannels(channels, tt.maxJobs)

			if err == nil {
				t.Fatal("Expected error for invalid maxJobs, got nil")
			}

			if assignments != nil {
				t.Errorf("Expected nil assignments on error, got %d assignments", len(assignments))
			}

			if err.Error() != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// TestAssignChannels_ChannelsExceedMaxJobs verifies validation when channels exceed available jobs
func TestAssignChannels_ChannelsExceedMaxJobs(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	tests := []struct {
		name          string
		channels      []string
		maxJobs       int
		expectedError string
	}{
		{
			name:          "5 channels with 3 jobs",
			channels:      []string{"ch1", "ch2", "ch3", "ch4", "ch5"},
			maxJobs:       3,
			expectedError: "channel count (5) exceeds available matrix jobs (3)",
		},
		{
			name:          "10 channels with 5 jobs",
			channels:      []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10"},
			maxJobs:       5,
			expectedError: "channel count (10) exceeds available matrix jobs (5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignments, err := mc.AssignChannels(tt.channels, tt.maxJobs)

			if err == nil {
				t.Fatal("Expected error when channels exceed maxJobs, got nil")
			}

			if assignments != nil {
				t.Errorf("Expected nil assignments on error, got %d assignments", len(assignments))
			}

			if err.Error() != tt.expectedError {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// TestAssignChannels_EmptyChannels verifies handling of empty channel list
func TestAssignChannels_EmptyChannels(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	assignments, err := mc.AssignChannels([]string{}, 5)

	if err != nil {
		t.Fatalf("AssignChannels returned unexpected error for empty channels: %v", err)
	}

	if len(assignments) != 0 {
		t.Errorf("Expected 0 assignments for empty channels, got %d", len(assignments))
	}
}

// TestFormatJobID verifies that formatJobID generates correct identifiers
func TestFormatJobID(t *testing.T) {
	tests := []struct {
		jobNumber int
		expected  string
	}{
		{1, "matrix-job-1"},
		{2, "matrix-job-2"},
		{5, "matrix-job-5"},
		{10, "matrix-job-10"},
		{15, "matrix-job-15"},
		{20, "matrix-job-20"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatJobID(tt.jobNumber)
			if result != tt.expected {
				t.Errorf("formatJobID(%d) = %s, expected %s", tt.jobNumber, result, tt.expected)
			}
		})
	}
}

// TestAssignChannels_OneChannelPerJob verifies exactly one channel per matrix job
func TestAssignChannels_OneChannelPerJob(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")
	channels := []string{"channel1", "channel2", "channel3", "channel4"}

	assignments, err := mc.AssignChannels(channels, 10)

	if err != nil {
		t.Fatalf("AssignChannels returned unexpected error: %v", err)
	}

	// Verify we have exactly as many assignments as channels
	if len(assignments) != len(channels) {
		t.Errorf("Expected %d assignments (one per channel), got %d", len(channels), len(assignments))
	}

	// Verify each assignment has a unique job ID and channel
	seenJobIDs := make(map[string]bool)
	seenChannels := make(map[string]bool)

	for _, assignment := range assignments {
		if seenJobIDs[assignment.JobID] {
			t.Errorf("JobID %s assigned to multiple channels", assignment.JobID)
		}
		seenJobIDs[assignment.JobID] = true

		if seenChannels[assignment.Channel] {
			t.Errorf("Channel %s assigned to multiple jobs", assignment.Channel)
		}
		seenChannels[assignment.Channel] = true
	}
}

// TestRegisterJob_ValidInput verifies that RegisterJob correctly adds jobs to the registry
func TestRegisterJob_ValidInput(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	tests := []struct {
		name    string
		jobID   string
		channel string
	}{
		{
			name:    "register first job",
			jobID:   "matrix-job-1",
			channel: "channel1",
		},
		{
			name:    "register second job",
			jobID:   "matrix-job-2",
			channel: "channel2",
		},
		{
			name:    "register job with long channel name",
			jobID:   "matrix-job-5",
			channel: "very_long_channel_username_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mc.RegisterJob(tt.jobID, tt.channel)

			if err != nil {
				t.Fatalf("RegisterJob returned unexpected error: %v", err)
			}

			// Verify job is in registry
			mc.registryMu.RLock()
			job, exists := mc.jobRegistry[tt.jobID]
			mc.registryMu.RUnlock()

			if !exists {
				t.Fatalf("Job %s not found in registry after registration", tt.jobID)
			}

			if job.JobID != tt.jobID {
				t.Errorf("Expected JobID %s, got %s", tt.jobID, job.JobID)
			}

			if job.Channel != tt.channel {
				t.Errorf("Expected Channel %s, got %s", tt.channel, job.Channel)
			}

			if job.Status != "starting" {
				t.Errorf("Expected Status 'starting', got '%s'", job.Status)
			}

			if job.StartTime.IsZero() {
				t.Error("StartTime should not be zero")
			}
		})
	}
}

// TestRegisterJob_EmptyJobID verifies validation when jobID is empty
func TestRegisterJob_EmptyJobID(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	err := mc.RegisterJob("", "channel1")

	if err == nil {
		t.Fatal("Expected error for empty jobID, got nil")
	}

	expectedError := "jobID cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// Verify registry is still empty
	if len(mc.jobRegistry) != 0 {
		t.Errorf("Expected empty registry, got %d entries", len(mc.jobRegistry))
	}
}

// TestRegisterJob_EmptyChannel verifies validation when channel is empty
func TestRegisterJob_EmptyChannel(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	err := mc.RegisterJob("matrix-job-1", "")

	if err == nil {
		t.Fatal("Expected error for empty channel, got nil")
	}

	expectedError := "channel cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// Verify registry is still empty
	if len(mc.jobRegistry) != 0 {
		t.Errorf("Expected empty registry, got %d entries", len(mc.jobRegistry))
	}
}

// TestRegisterJob_MultipleJobs verifies registering multiple jobs
func TestRegisterJob_MultipleJobs(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobs := []struct {
		jobID   string
		channel string
	}{
		{"matrix-job-1", "channel1"},
		{"matrix-job-2", "channel2"},
		{"matrix-job-3", "channel3"},
		{"matrix-job-4", "channel4"},
		{"matrix-job-5", "channel5"},
	}

	// Register all jobs
	for _, job := range jobs {
		err := mc.RegisterJob(job.jobID, job.channel)
		if err != nil {
			t.Fatalf("RegisterJob(%s, %s) returned unexpected error: %v", job.jobID, job.channel, err)
		}
	}

	// Verify all jobs are in registry
	mc.registryMu.RLock()
	registrySize := len(mc.jobRegistry)
	mc.registryMu.RUnlock()

	if registrySize != len(jobs) {
		t.Errorf("Expected %d jobs in registry, got %d", len(jobs), registrySize)
	}

	// Verify each job has correct data
	for _, job := range jobs {
		mc.registryMu.RLock()
		registeredJob, exists := mc.jobRegistry[job.jobID]
		mc.registryMu.RUnlock()

		if !exists {
			t.Errorf("Job %s not found in registry", job.jobID)
			continue
		}

		if registeredJob.Channel != job.channel {
			t.Errorf("Job %s: expected channel %s, got %s", job.jobID, job.channel, registeredJob.Channel)
		}
	}
}

// TestUnregisterJob_ValidInput verifies that UnregisterJob correctly removes jobs
func TestUnregisterJob_ValidInput(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register a job first
	jobID := "matrix-job-1"
	channel := "channel1"
	err := mc.RegisterJob(jobID, channel)
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	// Verify job is registered
	mc.registryMu.RLock()
	_, exists := mc.jobRegistry[jobID]
	mc.registryMu.RUnlock()

	if !exists {
		t.Fatal("Job should be registered before unregistering")
	}

	// Unregister the job
	err = mc.UnregisterJob(jobID)
	if err != nil {
		t.Fatalf("UnregisterJob returned unexpected error: %v", err)
	}

	// Verify job is removed from registry
	mc.registryMu.RLock()
	_, exists = mc.jobRegistry[jobID]
	mc.registryMu.RUnlock()

	if exists {
		t.Error("Job should be removed from registry after unregistering")
	}
}

// TestUnregisterJob_EmptyJobID verifies validation when jobID is empty
func TestUnregisterJob_EmptyJobID(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	err := mc.UnregisterJob("")

	if err == nil {
		t.Fatal("Expected error for empty jobID, got nil")
	}

	expectedError := "jobID cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestUnregisterJob_NotFound verifies error when job is not in registry
func TestUnregisterJob_NotFound(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobID := "matrix-job-99"
	err := mc.UnregisterJob(jobID)

	if err == nil {
		t.Fatal("Expected error for non-existent job, got nil")
	}

	expectedError := "job matrix-job-99 not found in registry"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestUnregisterJob_MultipleJobs verifies unregistering specific jobs
func TestUnregisterJob_MultipleJobs(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register multiple jobs
	jobs := []string{"matrix-job-1", "matrix-job-2", "matrix-job-3"}
	for _, jobID := range jobs {
		err := mc.RegisterJob(jobID, "channel-"+jobID)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	// Unregister the middle job
	err := mc.UnregisterJob("matrix-job-2")
	if err != nil {
		t.Fatalf("UnregisterJob failed: %v", err)
	}

	// Verify only matrix-job-2 is removed
	mc.registryMu.RLock()
	_, exists1 := mc.jobRegistry["matrix-job-1"]
	_, exists2 := mc.jobRegistry["matrix-job-2"]
	_, exists3 := mc.jobRegistry["matrix-job-3"]
	registrySize := len(mc.jobRegistry)
	mc.registryMu.RUnlock()

	if !exists1 {
		t.Error("matrix-job-1 should still be in registry")
	}

	if exists2 {
		t.Error("matrix-job-2 should be removed from registry")
	}

	if !exists3 {
		t.Error("matrix-job-3 should still be in registry")
	}

	if registrySize != 2 {
		t.Errorf("Expected 2 jobs in registry, got %d", registrySize)
	}
}

// TestGetActiveJobs_EmptyRegistry verifies GetActiveJobs with no registered jobs
func TestGetActiveJobs_EmptyRegistry(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobs := mc.GetActiveJobs()

	if jobs == nil {
		t.Fatal("GetActiveJobs should return empty slice, not nil")
	}

	if len(jobs) != 0 {
		t.Errorf("Expected 0 active jobs, got %d", len(jobs))
	}
}

// TestGetActiveJobs_SingleJob verifies GetActiveJobs with one registered job
func TestGetActiveJobs_SingleJob(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobID := "matrix-job-1"
	channel := "channel1"
	err := mc.RegisterJob(jobID, channel)
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	jobs := mc.GetActiveJobs()

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 active job, got %d", len(jobs))
	}

	job := jobs[0]
	if job.JobID != jobID {
		t.Errorf("Expected JobID %s, got %s", jobID, job.JobID)
	}

	if job.Channel != channel {
		t.Errorf("Expected Channel %s, got %s", channel, job.Channel)
	}

	if job.Status != "starting" {
		t.Errorf("Expected Status 'starting', got '%s'", job.Status)
	}
}

// TestGetActiveJobs_MultipleJobs verifies GetActiveJobs with multiple registered jobs
func TestGetActiveJobs_MultipleJobs(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	expectedJobs := []struct {
		jobID   string
		channel string
	}{
		{"matrix-job-1", "channel1"},
		{"matrix-job-2", "channel2"},
		{"matrix-job-3", "channel3"},
		{"matrix-job-4", "channel4"},
	}

	// Register all jobs
	for _, job := range expectedJobs {
		err := mc.RegisterJob(job.jobID, job.channel)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	jobs := mc.GetActiveJobs()

	if len(jobs) != len(expectedJobs) {
		t.Fatalf("Expected %d active jobs, got %d", len(expectedJobs), len(jobs))
	}

	// Create a map for easy lookup
	jobMap := make(map[string]MatrixJobInfo)
	for _, job := range jobs {
		jobMap[job.JobID] = job
	}

	// Verify all expected jobs are present
	for _, expected := range expectedJobs {
		job, exists := jobMap[expected.jobID]
		if !exists {
			t.Errorf("Expected job %s not found in active jobs", expected.jobID)
			continue
		}

		if job.Channel != expected.channel {
			t.Errorf("Job %s: expected channel %s, got %s", expected.jobID, expected.channel, job.Channel)
		}
	}
}

// TestGetActiveJobs_ReturnsCopy verifies that GetActiveJobs returns a copy
func TestGetActiveJobs_ReturnsCopy(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register a job
	err := mc.RegisterJob("matrix-job-1", "channel1")
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	// Get active jobs
	jobs1 := mc.GetActiveJobs()

	// Modify the returned slice
	if len(jobs1) > 0 {
		jobs1[0].Status = "modified"
	}

	// Get active jobs again
	jobs2 := mc.GetActiveJobs()

	// Verify the modification didn't affect the registry
	if len(jobs2) > 0 && jobs2[0].Status == "modified" {
		t.Error("GetActiveJobs should return a copy, not a reference to internal data")
	}

	if len(jobs2) > 0 && jobs2[0].Status != "starting" {
		t.Errorf("Expected Status 'starting', got '%s'", jobs2[0].Status)
	}
}

// TestGetActiveJobs_AfterUnregister verifies GetActiveJobs after unregistering jobs
func TestGetActiveJobs_AfterUnregister(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register multiple jobs
	jobs := []string{"matrix-job-1", "matrix-job-2", "matrix-job-3"}
	for _, jobID := range jobs {
		err := mc.RegisterJob(jobID, "channel-"+jobID)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	// Verify all jobs are active
	activeJobs := mc.GetActiveJobs()
	if len(activeJobs) != 3 {
		t.Fatalf("Expected 3 active jobs, got %d", len(activeJobs))
	}

	// Unregister one job
	err := mc.UnregisterJob("matrix-job-2")
	if err != nil {
		t.Fatalf("UnregisterJob failed: %v", err)
	}

	// Verify only 2 jobs are active
	activeJobs = mc.GetActiveJobs()
	if len(activeJobs) != 2 {
		t.Fatalf("Expected 2 active jobs after unregister, got %d", len(activeJobs))
	}

	// Verify the correct jobs remain
	jobMap := make(map[string]bool)
	for _, job := range activeJobs {
		jobMap[job.JobID] = true
	}

	if !jobMap["matrix-job-1"] {
		t.Error("matrix-job-1 should still be active")
	}

	if jobMap["matrix-job-2"] {
		t.Error("matrix-job-2 should not be active")
	}

	if !jobMap["matrix-job-3"] {
		t.Error("matrix-job-3 should still be active")
	}
}

// TestJobRegistry_ThreadSafety verifies thread-safe access to job registry
func TestJobRegistry_ThreadSafety(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Use a channel to synchronize goroutines
	done := make(chan bool)
	numGoroutines := 10

	// Concurrently register jobs
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			jobID := fmt.Sprintf("matrix-job-%d", id)
			channel := fmt.Sprintf("channel-%d", id)
			err := mc.RegisterJob(jobID, channel)
			if err != nil {
				t.Errorf("RegisterJob failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all registrations to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all jobs are registered
	jobs := mc.GetActiveJobs()
	if len(jobs) != numGoroutines {
		t.Errorf("Expected %d jobs, got %d", numGoroutines, len(jobs))
	}

	// Concurrently read active jobs
	for i := 0; i < numGoroutines; i++ {
		go func() {
			jobs := mc.GetActiveJobs()
			if len(jobs) != numGoroutines {
				t.Errorf("Expected %d jobs, got %d", numGoroutines, len(jobs))
			}
			done <- true
		}()
	}

	// Wait for all reads to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrently unregister jobs
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			jobID := fmt.Sprintf("matrix-job-%d", id)
			err := mc.UnregisterJob(jobID)
			if err != nil {
				t.Errorf("UnregisterJob failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all unregistrations to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all jobs are unregistered
	jobs = mc.GetActiveJobs()
	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs after unregister, got %d", len(jobs))
	}
}

// TestUpdateJobActivity_ValidInput verifies that UpdateJobActivity updates the timestamp
func TestUpdateJobActivity_ValidInput(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobID := "matrix-job-1"
	channel := "channel1"

	// Register a job
	err := mc.RegisterJob(jobID, channel)
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	// Get initial last activity time
	mc.registryMu.RLock()
	initialActivity := mc.jobRegistry[jobID].LastActivity
	mc.registryMu.RUnlock()

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update job activity
	err = mc.UpdateJobActivity(jobID)
	if err != nil {
		t.Fatalf("UpdateJobActivity returned unexpected error: %v", err)
	}

	// Get updated last activity time
	mc.registryMu.RLock()
	updatedActivity := mc.jobRegistry[jobID].LastActivity
	mc.registryMu.RUnlock()

	// Verify timestamp was updated
	if !updatedActivity.After(initialActivity) {
		t.Errorf("LastActivity should be updated, initial: %v, updated: %v", initialActivity, updatedActivity)
	}
}

// TestUpdateJobActivity_EmptyJobID verifies validation when jobID is empty
func TestUpdateJobActivity_EmptyJobID(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	err := mc.UpdateJobActivity("")

	if err == nil {
		t.Fatal("Expected error for empty jobID, got nil")
	}

	expectedError := "jobID cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestUpdateJobActivity_NotFound verifies error when job is not in registry
func TestUpdateJobActivity_NotFound(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobID := "matrix-job-99"
	err := mc.UpdateJobActivity(jobID)

	if err == nil {
		t.Fatal("Expected error for non-existent job, got nil")
	}

	expectedError := "job matrix-job-99 not found in registry"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestUpdateJobActivity_MultipleUpdates verifies multiple activity updates
func TestUpdateJobActivity_MultipleUpdates(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobID := "matrix-job-1"
	channel := "channel1"

	// Register a job
	err := mc.RegisterJob(jobID, channel)
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	// Perform multiple updates
	var previousActivity time.Time
	for i := 0; i < 5; i++ {
		time.Sleep(5 * time.Millisecond)

		err = mc.UpdateJobActivity(jobID)
		if err != nil {
			t.Fatalf("UpdateJobActivity failed on iteration %d: %v", i, err)
		}

		mc.registryMu.RLock()
		currentActivity := mc.jobRegistry[jobID].LastActivity
		mc.registryMu.RUnlock()

		if i > 0 && !currentActivity.After(previousActivity) {
			t.Errorf("Iteration %d: LastActivity should be updated, previous: %v, current: %v", i, previousActivity, currentActivity)
		}

		previousActivity = currentActivity
	}
}

// TestDetectFailedJobs_NoJobs verifies DetectFailedJobs with empty registry
func TestDetectFailedJobs_NoJobs(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	failedJobs := mc.DetectFailedJobs()

	if failedJobs == nil {
		t.Fatal("DetectFailedJobs should return empty slice, not nil")
	}

	if len(failedJobs) != 0 {
		t.Errorf("Expected 0 failed jobs, got %d", len(failedJobs))
	}
}

// TestDetectFailedJobs_AllJobsActive verifies no failed jobs when all are active
func TestDetectFailedJobs_AllJobsActive(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register multiple jobs
	jobs := []string{"matrix-job-1", "matrix-job-2", "matrix-job-3"}
	for _, jobID := range jobs {
		err := mc.RegisterJob(jobID, "channel-"+jobID)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	// All jobs were just registered, so they should be active
	failedJobs := mc.DetectFailedJobs()

	if len(failedJobs) != 0 {
		t.Errorf("Expected 0 failed jobs, got %d: %v", len(failedJobs), failedJobs)
	}
}

// TestDetectFailedJobs_SomeJobsStale verifies detection of stale jobs
func TestDetectFailedJobs_SomeJobsStale(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register multiple jobs
	jobs := []string{"matrix-job-1", "matrix-job-2", "matrix-job-3"}
	for _, jobID := range jobs {
		err := mc.RegisterJob(jobID, "channel-"+jobID)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	// Manually set LastActivity for some jobs to be old (simulate stale jobs)
	mc.registryMu.Lock()
	job1 := mc.jobRegistry["matrix-job-1"]
	job1.LastActivity = time.Now().Add(-15 * time.Minute) // 15 minutes ago (stale)
	mc.jobRegistry["matrix-job-1"] = job1

	job3 := mc.jobRegistry["matrix-job-3"]
	job3.LastActivity = time.Now().Add(-20 * time.Minute) // 20 minutes ago (stale)
	mc.jobRegistry["matrix-job-3"] = job3
	mc.registryMu.Unlock()

	// matrix-job-2 remains active (just registered)

	// Detect failed jobs with 10 minute timeout
	failedJobs := mc.DetectFailedJobsWithTimeout(10 * time.Minute)

	if len(failedJobs) != 2 {
		t.Fatalf("Expected 2 failed jobs, got %d: %v", len(failedJobs), failedJobs)
	}

	// Verify the correct jobs are detected as failed
	failedJobMap := make(map[string]bool)
	for _, jobID := range failedJobs {
		failedJobMap[jobID] = true
	}

	if !failedJobMap["matrix-job-1"] {
		t.Error("matrix-job-1 should be detected as failed")
	}

	if failedJobMap["matrix-job-2"] {
		t.Error("matrix-job-2 should not be detected as failed")
	}

	if !failedJobMap["matrix-job-3"] {
		t.Error("matrix-job-3 should be detected as failed")
	}
}

// TestDetectFailedJobs_AllJobsStale verifies detection when all jobs are stale
func TestDetectFailedJobs_AllJobsStale(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register multiple jobs
	jobs := []string{"matrix-job-1", "matrix-job-2", "matrix-job-3"}
	for _, jobID := range jobs {
		err := mc.RegisterJob(jobID, "channel-"+jobID)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	// Manually set LastActivity for all jobs to be old
	mc.registryMu.Lock()
	for jobID := range mc.jobRegistry {
		job := mc.jobRegistry[jobID]
		job.LastActivity = time.Now().Add(-15 * time.Minute)
		mc.jobRegistry[jobID] = job
	}
	mc.registryMu.Unlock()

	// Detect failed jobs
	failedJobs := mc.DetectFailedJobsWithTimeout(10 * time.Minute)

	if len(failedJobs) != len(jobs) {
		t.Errorf("Expected %d failed jobs, got %d: %v", len(jobs), len(failedJobs), failedJobs)
	}
}

// TestDetectFailedJobs_CustomTimeout verifies custom timeout parameter
func TestDetectFailedJobs_CustomTimeout(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register a job
	jobID := "matrix-job-1"
	err := mc.RegisterJob(jobID, "channel1")
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	// Set LastActivity to 3 minutes ago
	mc.registryMu.Lock()
	job := mc.jobRegistry[jobID]
	job.LastActivity = time.Now().Add(-3 * time.Minute)
	mc.jobRegistry[jobID] = job
	mc.registryMu.Unlock()

	// With 5 minute timeout, job should not be failed
	failedJobs := mc.DetectFailedJobsWithTimeout(5 * time.Minute)
	if len(failedJobs) != 0 {
		t.Errorf("Expected 0 failed jobs with 5 minute timeout, got %d", len(failedJobs))
	}

	// With 2 minute timeout, job should be failed
	failedJobs = mc.DetectFailedJobsWithTimeout(2 * time.Minute)
	if len(failedJobs) != 1 {
		t.Errorf("Expected 1 failed job with 2 minute timeout, got %d", len(failedJobs))
	}

	if len(failedJobs) > 0 && failedJobs[0] != jobID {
		t.Errorf("Expected failed job %s, got %s", jobID, failedJobs[0])
	}
}

// TestDetectFailedJobs_BoundaryCondition verifies detection at exact timeout boundary
func TestDetectFailedJobs_BoundaryCondition(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register a job
	jobID := "matrix-job-1"
	err := mc.RegisterJob(jobID, "channel1")
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	timeout := 5 * time.Minute

	// Set LastActivity to exactly timeout duration ago
	mc.registryMu.Lock()
	job := mc.jobRegistry[jobID]
	job.LastActivity = time.Now().Add(-timeout)
	mc.jobRegistry[jobID] = job
	mc.registryMu.Unlock()

	// Job at exact timeout boundary should not be considered failed (> not >=)
	// Wait a tiny bit to ensure we're past the boundary
	time.Sleep(10 * time.Millisecond)

	failedJobs := mc.DetectFailedJobsWithTimeout(timeout)

	// Should be detected as failed since we're now past the timeout
	if len(failedJobs) != 1 {
		t.Errorf("Expected 1 failed job at boundary, got %d", len(failedJobs))
	}
}

// TestDetectFailedJobs_AfterActivityUpdate verifies jobs are not failed after activity update
func TestDetectFailedJobs_AfterActivityUpdate(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register a job
	jobID := "matrix-job-1"
	err := mc.RegisterJob(jobID, "channel1")
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	// Set LastActivity to be old (stale)
	mc.registryMu.Lock()
	job := mc.jobRegistry[jobID]
	job.LastActivity = time.Now().Add(-15 * time.Minute)
	mc.jobRegistry[jobID] = job
	mc.registryMu.Unlock()

	// Verify job is detected as failed
	failedJobs := mc.DetectFailedJobsWithTimeout(10 * time.Minute)
	if len(failedJobs) != 1 {
		t.Fatalf("Expected 1 failed job before update, got %d", len(failedJobs))
	}

	// Update job activity
	err = mc.UpdateJobActivity(jobID)
	if err != nil {
		t.Fatalf("UpdateJobActivity failed: %v", err)
	}

	// Verify job is no longer detected as failed
	failedJobs = mc.DetectFailedJobsWithTimeout(10 * time.Minute)
	if len(failedJobs) != 0 {
		t.Errorf("Expected 0 failed jobs after activity update, got %d: %v", len(failedJobs), failedJobs)
	}
}

// TestDetectFailedJobs_ThreadSafety verifies thread-safe detection
func TestDetectFailedJobs_ThreadSafety(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	// Register multiple jobs
	numJobs := 10
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("matrix-job-%d", i)
		err := mc.RegisterJob(jobID, "channel-"+jobID)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}
	}

	// Concurrently detect failed jobs and update activity
	done := make(chan bool)
	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			if id%2 == 0 {
				// Even goroutines detect failed jobs
				_ = mc.DetectFailedJobs()
			} else {
				// Odd goroutines update activity
				jobID := fmt.Sprintf("matrix-job-%d", id%numJobs)
				_ = mc.UpdateJobActivity(jobID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify registry is still consistent
	jobs := mc.GetActiveJobs()
	if len(jobs) != numJobs {
		t.Errorf("Expected %d jobs after concurrent operations, got %d", numJobs, len(jobs))
	}
}

// TestRegisterJob_SetsLastActivity verifies that RegisterJob sets LastActivity
func TestRegisterJob_SetsLastActivity(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	jobID := "matrix-job-1"
	channel := "channel1"

	beforeRegister := time.Now()
	err := mc.RegisterJob(jobID, channel)
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}
	afterRegister := time.Now()

	mc.registryMu.RLock()
	job := mc.jobRegistry[jobID]
	mc.registryMu.RUnlock()

	// Verify LastActivity is set and within expected time range
	if job.LastActivity.IsZero() {
		t.Error("LastActivity should be set when job is registered")
	}

	if job.LastActivity.Before(beforeRegister) || job.LastActivity.After(afterRegister) {
		t.Errorf("LastActivity (%v) should be between %v and %v", job.LastActivity, beforeRegister, afterRegister)
	}

	// Verify LastActivity equals StartTime at registration
	if !job.LastActivity.Equal(job.StartTime) {
		t.Errorf("LastActivity (%v) should equal StartTime (%v) at registration", job.LastActivity, job.StartTime)
	}
}

// TestGetJobCacheKey verifies that GetJobCacheKey generates correct cache keys
func TestGetJobCacheKey(t *testing.T) {
	sessionID := "test-session-abc123"
	mc := NewMatrixCoordinator(sessionID)

	tests := []struct {
		name         string
		matrixJobID  string
		expectedKey  string
	}{
		{
			name:        "job 1",
			matrixJobID: "matrix-job-1",
			expectedKey: "state-test-session-abc123-matrix-job-1",
		},
		{
			name:        "job 5",
			matrixJobID: "matrix-job-5",
			expectedKey: "state-test-session-abc123-matrix-job-5",
		},
		{
			name:        "job 10",
			matrixJobID: "matrix-job-10",
			expectedKey: "state-test-session-abc123-matrix-job-10",
		},
		{
			name:        "job 20",
			matrixJobID: "matrix-job-20",
			expectedKey: "state-test-session-abc123-matrix-job-20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheKey := mc.GetJobCacheKey(tt.matrixJobID)

			if cacheKey != tt.expectedKey {
				t.Errorf("Expected cache key '%s', got '%s'", tt.expectedKey, cacheKey)
			}
		})
	}
}

// TestGetJobCacheKey_Uniqueness verifies that each job gets a unique cache key
func TestGetJobCacheKey_Uniqueness(t *testing.T) {
	sessionID := "test-session-xyz"
	mc := NewMatrixCoordinator(sessionID)

	// Generate cache keys for multiple jobs
	numJobs := 20
	cacheKeys := make(map[string]bool)

	for i := 1; i <= numJobs; i++ {
		matrixJobID := fmt.Sprintf("matrix-job-%d", i)
		cacheKey := mc.GetJobCacheKey(matrixJobID)

		// Verify cache key is unique
		if cacheKeys[cacheKey] {
			t.Errorf("Duplicate cache key found: %s", cacheKey)
		}
		cacheKeys[cacheKey] = true

		// Verify cache key contains session ID and job ID
		expectedSubstring1 := sessionID
		expectedSubstring2 := matrixJobID
		if !contains(cacheKey, expectedSubstring1) {
			t.Errorf("Cache key '%s' should contain session ID '%s'", cacheKey, expectedSubstring1)
		}
		if !contains(cacheKey, expectedSubstring2) {
			t.Errorf("Cache key '%s' should contain job ID '%s'", cacheKey, expectedSubstring2)
		}
	}

	// Verify we have exactly numJobs unique keys
	if len(cacheKeys) != numJobs {
		t.Errorf("Expected %d unique cache keys, got %d", numJobs, len(cacheKeys))
	}
}

// TestGetJobCacheKey_DifferentSessions verifies cache keys differ across sessions
func TestGetJobCacheKey_DifferentSessions(t *testing.T) {
	mc1 := NewMatrixCoordinator("session-1")
	mc2 := NewMatrixCoordinator("session-2")

	matrixJobID := "matrix-job-1"

	key1 := mc1.GetJobCacheKey(matrixJobID)
	key2 := mc2.GetJobCacheKey(matrixJobID)

	// Verify keys are different for different sessions
	if key1 == key2 {
		t.Errorf("Cache keys should differ for different sessions, both got: %s", key1)
	}

	// Verify each key contains its respective session ID
	if !contains(key1, "session-1") {
		t.Errorf("Key1 '%s' should contain 'session-1'", key1)
	}
	if !contains(key2, "session-2") {
		t.Errorf("Key2 '%s' should contain 'session-2'", key2)
	}
}

// TestGetSharedConfigCacheKey verifies the shared configuration cache key
func TestGetSharedConfigCacheKey(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	cacheKey := mc.GetSharedConfigCacheKey()

	expectedKey := "shared-config-latest"
	if cacheKey != expectedKey {
		t.Errorf("Expected shared config cache key '%s', got '%s'", expectedKey, cacheKey)
	}
}

// TestGetSharedConfigCacheKey_SameAcrossSessions verifies shared key is consistent
func TestGetSharedConfigCacheKey_SameAcrossSessions(t *testing.T) {
	mc1 := NewMatrixCoordinator("session-1")
	mc2 := NewMatrixCoordinator("session-2")
	mc3 := NewMatrixCoordinator("session-3")

	key1 := mc1.GetSharedConfigCacheKey()
	key2 := mc2.GetSharedConfigCacheKey()
	key3 := mc3.GetSharedConfigCacheKey()

	// Verify all coordinators return the same shared config key
	if key1 != key2 || key2 != key3 {
		t.Errorf("Shared config keys should be identical across sessions, got: %s, %s, %s", key1, key2, key3)
	}

	expectedKey := "shared-config-latest"
	if key1 != expectedKey {
		t.Errorf("Expected shared config cache key '%s', got '%s'", expectedKey, key1)
	}
}

// TestGetSharedConfigCacheKey_DifferentFromJobKeys verifies shared key differs from job keys
func TestGetSharedConfigCacheKey_DifferentFromJobKeys(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	sharedKey := mc.GetSharedConfigCacheKey()

	// Generate job cache keys and verify they differ from shared key
	for i := 1; i <= 20; i++ {
		matrixJobID := fmt.Sprintf("matrix-job-%d", i)
		jobKey := mc.GetJobCacheKey(matrixJobID)

		if jobKey == sharedKey {
			t.Errorf("Job cache key '%s' should differ from shared config key '%s'", jobKey, sharedKey)
		}
	}
}

// TestValidateJobCacheKey_ValidKeys verifies validation of correct cache keys
func TestValidateJobCacheKey_ValidKeys(t *testing.T) {
	sessionID := "test-session-validation"
	mc := NewMatrixCoordinator(sessionID)

	tests := []struct {
		name        string
		matrixJobID string
	}{
		{"job 1", "matrix-job-1"},
		{"job 5", "matrix-job-5"},
		{"job 10", "matrix-job-10"},
		{"job 20", "matrix-job-20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the correct cache key
			correctKey := mc.GetJobCacheKey(tt.matrixJobID)

			// Validate it
			isValid := mc.ValidateJobCacheKey(tt.matrixJobID, correctKey)

			if !isValid {
				t.Errorf("ValidateJobCacheKey should return true for correct key '%s'", correctKey)
			}
		})
	}
}

// TestValidateJobCacheKey_InvalidKeys verifies validation rejects incorrect cache keys
func TestValidateJobCacheKey_InvalidKeys(t *testing.T) {
	sessionID := "test-session-validation"
	mc := NewMatrixCoordinator(sessionID)

	tests := []struct {
		name        string
		matrixJobID string
		wrongKey    string
	}{
		{
			name:        "wrong job ID",
			matrixJobID: "matrix-job-1",
			wrongKey:    "state-test-session-validation-matrix-job-2",
		},
		{
			name:        "wrong session ID",
			matrixJobID: "matrix-job-1",
			wrongKey:    "state-different-session-matrix-job-1",
		},
		{
			name:        "wrong format",
			matrixJobID: "matrix-job-1",
			wrongKey:    "invalid-cache-key",
		},
		{
			name:        "shared config key",
			matrixJobID: "matrix-job-1",
			wrongKey:    "shared-config-latest",
		},
		{
			name:        "empty key",
			matrixJobID: "matrix-job-1",
			wrongKey:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := mc.ValidateJobCacheKey(tt.matrixJobID, tt.wrongKey)

			if isValid {
				t.Errorf("ValidateJobCacheKey should return false for incorrect key '%s'", tt.wrongKey)
			}
		})
	}
}

// TestValidateJobCacheKey_CrossJobValidation verifies jobs can't use each other's keys
func TestValidateJobCacheKey_CrossJobValidation(t *testing.T) {
	mc := NewMatrixCoordinator("test-session")

	job1ID := "matrix-job-1"
	job2ID := "matrix-job-2"

	job1Key := mc.GetJobCacheKey(job1ID)
	job2Key := mc.GetJobCacheKey(job2ID)

	// Verify job1's key is valid for job1
	if !mc.ValidateJobCacheKey(job1ID, job1Key) {
		t.Error("Job1's key should be valid for job1")
	}

	// Verify job2's key is valid for job2
	if !mc.ValidateJobCacheKey(job2ID, job2Key) {
		t.Error("Job2's key should be valid for job2")
	}

	// Verify job1's key is NOT valid for job2
	if mc.ValidateJobCacheKey(job2ID, job1Key) {
		t.Error("Job1's key should not be valid for job2")
	}

	// Verify job2's key is NOT valid for job1
	if mc.ValidateJobCacheKey(job1ID, job2Key) {
		t.Error("Job2's key should not be valid for job1")
	}
}

// TestCacheKeyManagement_Integration verifies cache key management workflow
func TestCacheKeyManagement_Integration(t *testing.T) {
	sessionID := "integration-test-session"
	mc := NewMatrixCoordinator(sessionID)

	// Simulate assigning channels to jobs
	channels := []string{"channel1", "channel2", "channel3"}
	assignments, err := mc.AssignChannels(channels, 5)
	if err != nil {
		t.Fatalf("AssignChannels failed: %v", err)
	}

	// For each assignment, verify cache key management
	for _, assignment := range assignments {
		// Register the job
		err := mc.RegisterJob(assignment.JobID, assignment.Channel)
		if err != nil {
			t.Fatalf("RegisterJob failed: %v", err)
		}

		// Get the job's cache key
		jobCacheKey := mc.GetJobCacheKey(assignment.JobID)

		// Verify the cache key is valid for this job
		if !mc.ValidateJobCacheKey(assignment.JobID, jobCacheKey) {
			t.Errorf("Cache key '%s' should be valid for job '%s'", jobCacheKey, assignment.JobID)
		}

		// Verify the cache key contains the session ID
		if !contains(jobCacheKey, sessionID) {
			t.Errorf("Cache key '%s' should contain session ID '%s'", jobCacheKey, sessionID)
		}

		// Verify the cache key contains the job ID
		if !contains(jobCacheKey, assignment.JobID) {
			t.Errorf("Cache key '%s' should contain job ID '%s'", jobCacheKey, assignment.JobID)
		}

		// Verify the shared config key is different from job key
		sharedKey := mc.GetSharedConfigCacheKey()
		if jobCacheKey == sharedKey {
			t.Errorf("Job cache key '%s' should differ from shared config key '%s'", jobCacheKey, sharedKey)
		}
	}

	// Verify all jobs have unique cache keys
	cacheKeys := make(map[string]bool)
	for _, assignment := range assignments {
		jobCacheKey := mc.GetJobCacheKey(assignment.JobID)
		if cacheKeys[jobCacheKey] {
			t.Errorf("Duplicate cache key found: %s", jobCacheKey)
		}
		cacheKeys[jobCacheKey] = true
	}

	if len(cacheKeys) != len(assignments) {
		t.Errorf("Expected %d unique cache keys, got %d", len(assignments), len(cacheKeys))
	}
}


// TestCacheKeyFormat verifies the cache key format matches specification
func TestCacheKeyFormat(t *testing.T) {
	sessionID := "session-12345"
	mc := NewMatrixCoordinator(sessionID)

	tests := []struct {
		name           string
		matrixJobID    string
		expectedPrefix string
		expectedFormat string
	}{
		{
			name:           "job 1",
			matrixJobID:    "matrix-job-1",
			expectedPrefix: "state-",
			expectedFormat: "state-session-12345-matrix-job-1",
		},
		{
			name:           "job 15",
			matrixJobID:    "matrix-job-15",
			expectedPrefix: "state-",
			expectedFormat: "state-session-12345-matrix-job-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheKey := mc.GetJobCacheKey(tt.matrixJobID)

			// Verify key starts with expected prefix
			if len(cacheKey) < len(tt.expectedPrefix) || cacheKey[:len(tt.expectedPrefix)] != tt.expectedPrefix {
				t.Errorf("Cache key '%s' should start with '%s'", cacheKey, tt.expectedPrefix)
			}

			// Verify key matches expected format
			if cacheKey != tt.expectedFormat {
				t.Errorf("Expected cache key format '%s', got '%s'", tt.expectedFormat, cacheKey)
			}
		})
	}
}


