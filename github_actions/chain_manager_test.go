package github_actions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateSessionID(t *testing.T) {
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")

	// Test that session ID is generated
	sessionID := cm.GenerateSessionID()
	if sessionID == "" {
		t.Fatal("GenerateSessionID returned empty string")
	}

	// Test that session ID has expected format: run-YYYYMMDD-HHMMSS-{hex}
	if !strings.HasPrefix(sessionID, "run-") {
		t.Errorf("Session ID should start with 'run-', got: %s", sessionID)
	}

	// Test that session ID is stored in ChainManager
	if cm.GetSessionID() != sessionID {
		t.Errorf("Session ID not stored correctly, expected %s, got %s", sessionID, cm.GetSessionID())
	}

	// Test uniqueness - generate multiple IDs
	cm2 := NewChainManager("test-token", "owner/repo", "workflow.yml")
	sessionID2 := cm2.GenerateSessionID()
	if sessionID == sessionID2 {
		t.Error("GenerateSessionID should generate unique IDs")
	}
}

func TestTriggerNextRun_Success(t *testing.T) {
	// Create a mock GitHub API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify URL path
		expectedPath := "/repos/owner/repo/actions/workflows/workflow.yml/dispatches"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("Expected Accept header 'application/vnd.github+json', got %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Verify payload
		var payload workflowDispatchPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		if payload.Ref != "main" {
			t.Errorf("Expected ref 'main', got %s", payload.Ref)
		}

		if payload.Inputs["channels"] != "channel1,channel2" {
			t.Errorf("Expected channels 'channel1,channel2', got %v", payload.Inputs["channels"])
		}

		if payload.Inputs["matrix_job_count"] != "5" {
			t.Errorf("Expected matrix_job_count '5', got %v", payload.Inputs["matrix_job_count"])
		}

		// Verify session_state is valid JSON
		stateJSON, ok := payload.Inputs["session_state"].(string)
		if !ok {
			t.Error("session_state should be a string")
		}
		var state SessionState
		if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
			t.Errorf("session_state is not valid JSON: %v", err)
		}

		// Return success response (204 No Content)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	// Create ChainManager with mock server URL
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()

	// Create test session state
	state := SessionState{
		SessionID:      cm.GetSessionID(),
		StartTime:      time.Now(),
		Channels:       []string{"channel1", "channel2"},
		MatrixJobCount: 5,
		Configuration:  map[string]interface{}{"key": "value"},
	}

	// Test TriggerNextRun
	ctx := context.Background()
	err := cm.TriggerNextRun(ctx, state)
	if err != nil {
		t.Fatalf("TriggerNextRun failed: %v", err)
	}

	// Verify that nextRunTriggered flag is set
	if !cm.IsNextRunTriggered() {
		t.Error("nextRunTriggered flag should be true after successful trigger")
	}
}

func TestTriggerNextRun_AlreadyTriggered(t *testing.T) {
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.GenerateSessionID()
	cm.nextRunTriggered = true

	state := SessionState{
		SessionID: cm.GetSessionID(),
		Channels:  []string{"channel1"},
	}

	ctx := context.Background()
	err := cm.TriggerNextRun(ctx, state)
	if err == nil {
		t.Error("TriggerNextRun should fail when already triggered")
	}
	if !strings.Contains(err.Error(), "already triggered") {
		t.Errorf("Error should mention 'already triggered', got: %v", err)
	}
}

func TestTriggerNextRun_APIError(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer mockServer.Close()

	cm := NewChainManager("invalid-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()

	state := SessionState{
		SessionID: cm.GetSessionID(),
		Channels:  []string{"channel1"},
	}

	ctx := context.Background()
	err := cm.TriggerNextRun(ctx, state)
	if err == nil {
		t.Error("TriggerNextRun should fail with invalid token")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Error should mention status 401, got: %v", err)
	}
}

func TestJoinChannels(t *testing.T) {
	tests := []struct {
		name     string
		channels []string
		expected string
	}{
		{
			name:     "empty slice",
			channels: []string{},
			expected: "",
		},
		{
			name:     "single channel",
			channels: []string{"channel1"},
			expected: "channel1",
		},
		{
			name:     "multiple channels",
			channels: []string{"channel1", "channel2", "channel3"},
			expected: "channel1,channel2,channel3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinChannels(tt.channels)
			if result != tt.expected {
				t.Errorf("joinChannels(%v) = %s, expected %s", tt.channels, result, tt.expected)
			}
		})
	}
}

func TestMonitorRuntime_TriggersAtThreshold(t *testing.T) {
	// Create a mock GitHub API server
	triggered := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	// Create ChainManager with mock server
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()
	
	// Set start time to 5.5 hours ago to trigger immediately
	cm.startTime = time.Now().Add(-19800 * time.Second)

	// Create state provider
	stateProvider := func() SessionState {
		return SessionState{
			SessionID:      cm.GetSessionID(),
			StartTime:      cm.GetStartTime(),
			Channels:       []string{"channel1"},
			MatrixJobCount: 1,
		}
	}

	// Run MonitorRuntime with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cm.MonitorRuntime(ctx, stateProvider)
	if err != nil {
		t.Fatalf("MonitorRuntime failed: %v", err)
	}

	// Verify that the next run was triggered
	if !triggered {
		t.Error("Expected next workflow run to be triggered")
	}
	if !cm.IsNextRunTriggered() {
		t.Error("nextRunTriggered flag should be true")
	}
}

func TestMonitorRuntime_WaitsUntilThreshold(t *testing.T) {
	// Create a mock GitHub API server
	triggered := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		triggered = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	// Create ChainManager with mock server
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()
	
	// Set start time to just before threshold (5.4 hours ago)
	// This should not trigger immediately
	cm.startTime = time.Now().Add(-19440 * time.Second)

	// Create state provider
	stateProvider := func() SessionState {
		return SessionState{
			SessionID:      cm.GetSessionID(),
			StartTime:      cm.GetStartTime(),
			Channels:       []string{"channel1"},
			MatrixJobCount: 1,
		}
	}

	// Run MonitorRuntime with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := cm.MonitorRuntime(ctx, stateProvider)
	
	// Should timeout before triggering
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
	if triggered {
		t.Error("Should not trigger before threshold")
	}
	if cm.IsNextRunTriggered() {
		t.Error("nextRunTriggered flag should be false")
	}
}

func TestMonitorRuntime_ContextCancellation(t *testing.T) {
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.GenerateSessionID()

	stateProvider := func() SessionState {
		return SessionState{
			SessionID: cm.GetSessionID(),
			Channels:  []string{"channel1"},
		}
	}

	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := cm.MonitorRuntime(ctx, stateProvider)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

func TestMonitorRuntime_TriggerFailure(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Internal server error"}`))
	}))
	defer mockServer.Close()

	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()
	
	// Set start time to 5.5 hours ago to trigger immediately
	cm.startTime = time.Now().Add(-19800 * time.Second)

	stateProvider := func() SessionState {
		return SessionState{
			SessionID: cm.GetSessionID(),
			Channels:  []string{"channel1"},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cm.MonitorRuntime(ctx, stateProvider)
	if err == nil {
		t.Error("Expected error when trigger fails")
	}
	if !strings.Contains(err.Error(), "failed to trigger next workflow run") {
		t.Errorf("Error should mention trigger failure, got: %v", err)
	}
}

func TestGetElapsedTime(t *testing.T) {
	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	
	// Set start time to 1 hour ago
	cm.startTime = time.Now().Add(-1 * time.Hour)
	
	elapsed := cm.GetElapsedTime()
	
	// Should be approximately 1 hour (allow some tolerance)
	if elapsed < 59*time.Minute || elapsed > 61*time.Minute {
		t.Errorf("Expected elapsed time around 1 hour, got: %v", elapsed)
	}
}

func TestRetryWithBackoff_SuccessFirstAttempt(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	operation := func() error {
		attempts++
		return nil // Success on first attempt
	}

	err := RetryWithBackoff(ctx, 3, operation)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got: %d", attempts)
	}
}

func TestRetryWithBackoff_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	operation := func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("temporary error")
		}
		return nil // Success on third attempt
	}

	start := time.Now()
	err := RetryWithBackoff(ctx, 3, operation)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}

	// Should have waited 1s + 2s = 3s total (with some tolerance)
	if elapsed < 2900*time.Millisecond || elapsed > 3500*time.Millisecond {
		t.Errorf("Expected elapsed time around 3s, got: %v", elapsed)
	}
}

func TestRetryWithBackoff_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	operation := func() error {
		attempts++
		return fmt.Errorf("persistent error")
	}

	err := RetryWithBackoff(ctx, 3, operation)
	if err == nil {
		t.Error("Expected error after all attempts fail")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
	if !strings.Contains(err.Error(), "operation failed after 3 attempts") {
		t.Errorf("Error should mention failed attempts, got: %v", err)
	}
	if !strings.Contains(err.Error(), "persistent error") {
		t.Errorf("Error should include original error, got: %v", err)
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	operation := func() error {
		attempts++
		if attempts == 1 {
			// Cancel context after first failure
			cancel()
		}
		return fmt.Errorf("error")
	}

	err := RetryWithBackoff(ctx, 3, operation)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
	if !strings.Contains(err.Error(), "retry cancelled") {
		t.Errorf("Error should mention cancellation, got: %v", err)
	}
	// Should only attempt once before cancellation
	if attempts != 1 {
		t.Errorf("Expected 1 attempt before cancellation, got: %d", attempts)
	}
}

func TestRetryWithBackoff_ZeroMaxAttempts(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	operation := func() error {
		attempts++
		return nil
	}

	err := RetryWithBackoff(ctx, 0, operation)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	// Should default to 1 attempt
	if attempts != 1 {
		t.Errorf("Expected 1 attempt with maxAttempts=0, got: %d", attempts)
	}
}

func TestTriggerNextRun_WithRetry(t *testing.T) {
	attempts := 0
	// Create a mock server that fails twice then succeeds
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail first two attempts
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"message": "Service temporarily unavailable"}`))
			return
		}
		// Succeed on third attempt
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()

	state := SessionState{
		SessionID:      cm.GetSessionID(),
		StartTime:      time.Now(),
		Channels:       []string{"channel1"},
		MatrixJobCount: 1,
	}

	ctx := context.Background()
	start := time.Now()
	err := cm.TriggerNextRun(ctx, state)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("TriggerNextRun should succeed after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
	if !cm.IsNextRunTriggered() {
		t.Error("nextRunTriggered flag should be true after successful trigger")
	}

	// Should have waited 1s + 2s = 3s total (with some tolerance)
	if elapsed < 2900*time.Millisecond || elapsed > 3500*time.Millisecond {
		t.Errorf("Expected elapsed time around 3s, got: %v", elapsed)
	}
}

func TestTriggerNextRun_RetriesExhausted(t *testing.T) {
	attempts := 0
	// Create a mock server that always fails
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"message": "Service unavailable"}`))
	}))
	defer mockServer.Close()

	cm := NewChainManager("test-token", "owner/repo", "workflow.yml")
	cm.httpClient.Transport = &mockTransport{baseURL: mockServer.URL}
	cm.GenerateSessionID()

	state := SessionState{
		SessionID: cm.GetSessionID(),
		Channels:  []string{"channel1"},
	}

	ctx := context.Background()
	err := cm.TriggerNextRun(ctx, state)

	if err == nil {
		t.Error("TriggerNextRun should fail after all retries exhausted")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
	if !strings.Contains(err.Error(), "operation failed after 3 attempts") {
		t.Errorf("Error should mention failed attempts, got: %v", err)
	}
	if cm.IsNextRunTriggered() {
		t.Error("nextRunTriggered flag should be false after failure")
	}
}

// mockTransport is a custom http.RoundTripper for testing
type mockTransport struct {
	baseURL string
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the GitHub API URL with our mock server URL
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.baseURL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}

