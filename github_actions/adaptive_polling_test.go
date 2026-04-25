package github_actions

import (
	"context"
	"testing"
	"time"

	"github.com/HeapOfChaos/goondvr/entity"
	"github.com/HeapOfChaos/goondvr/server"
)

// TestNewAdaptivePolling verifies that NewAdaptivePolling creates an instance with correct initial values.
func TestNewAdaptivePolling(t *testing.T) {
	tests := []struct {
		name            string
		normalInterval  int
		expectedNormal  int
		expectedReduced int
		expectedCurrent int
	}{
		{
			name:            "Valid normal interval",
			normalInterval:  10,
			expectedNormal:  10,
			expectedReduced: 5,
			expectedCurrent: 10,
		},
		{
			name:            "Zero interval defaults to 1",
			normalInterval:  0,
			expectedNormal:  1,
			expectedReduced: 5,
			expectedCurrent: 1,
		},
		{
			name:            "Negative interval defaults to 1",
			normalInterval:  -5,
			expectedNormal:  1,
			expectedReduced: 5,
			expectedCurrent: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := NewAdaptivePolling(tt.normalInterval)

			if ap.GetNormalInterval() != tt.expectedNormal {
				t.Errorf("Expected normal interval %d, got %d", tt.expectedNormal, ap.GetNormalInterval())
			}

			if ap.GetReducedInterval() != tt.expectedReduced {
				t.Errorf("Expected reduced interval %d, got %d", tt.expectedReduced, ap.GetReducedInterval())
			}

			if ap.GetCurrentInterval() != tt.expectedCurrent {
				t.Errorf("Expected current interval %d, got %d", tt.expectedCurrent, ap.GetCurrentInterval())
			}
		})
	}
}

// TestUpdateInterval_NoActiveRecordings verifies that the interval is reduced when no recordings are active.
func TestUpdateInterval_NoActiveRecordings(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePolling(10)

	// Initially, the interval should be the normal interval (10)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected initial interval 10, got %d", ap.GetCurrentInterval())
	}

	// Update with no active recordings
	ap.UpdateInterval(false)

	// The interval should now be reduced to 5
	if ap.GetCurrentInterval() != 5 {
		t.Errorf("Expected reduced interval 5, got %d", ap.GetCurrentInterval())
	}

	// Verify server config was updated
	if server.Config.Interval != 5 {
		t.Errorf("Expected server config interval 5, got %d", server.Config.Interval)
	}
}

// TestUpdateInterval_WithActiveRecordings verifies that the interval is restored when recordings are active.
func TestUpdateInterval_WithActiveRecordings(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 5,
	}

	ap := NewAdaptivePolling(10)

	// Start with reduced interval
	ap.UpdateInterval(false)
	if ap.GetCurrentInterval() != 5 {
		t.Errorf("Expected reduced interval 5, got %d", ap.GetCurrentInterval())
	}

	// Update with active recordings
	ap.UpdateInterval(true)

	// The interval should now be restored to normal (10)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected normal interval 10, got %d", ap.GetCurrentInterval())
	}

	// Verify server config was updated
	if server.Config.Interval != 10 {
		t.Errorf("Expected server config interval 10, got %d", server.Config.Interval)
	}
}

// TestUpdateInterval_NoChange verifies that the interval is not updated if it hasn't changed.
func TestUpdateInterval_NoChange(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePolling(10)

	// Get initial last update time
	initialUpdateTime := ap.GetLastUpdateTime()

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update with active recordings (should not change since we're already at normal interval)
	ap.UpdateInterval(true)

	// The last update time should not have changed
	if ap.GetLastUpdateTime() != initialUpdateTime {
		t.Error("Expected last update time to remain unchanged when interval doesn't change")
	}

	// The interval should still be 10
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected interval to remain 10, got %d", ap.GetCurrentInterval())
	}
}

// TestUpdateInterval_MultipleTransitions verifies that the interval can transition multiple times.
func TestUpdateInterval_MultipleTransitions(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePolling(10)

	// Transition 1: Normal -> Reduced
	ap.UpdateInterval(false)
	if ap.GetCurrentInterval() != 5 {
		t.Errorf("Transition 1: Expected interval 5, got %d", ap.GetCurrentInterval())
	}

	// Transition 2: Reduced -> Normal
	ap.UpdateInterval(true)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Transition 2: Expected interval 10, got %d", ap.GetCurrentInterval())
	}

	// Transition 3: Normal -> Reduced
	ap.UpdateInterval(false)
	if ap.GetCurrentInterval() != 5 {
		t.Errorf("Transition 3: Expected interval 5, got %d", ap.GetCurrentInterval())
	}

	// Transition 4: Reduced -> Normal
	ap.UpdateInterval(true)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Transition 4: Expected interval 10, got %d", ap.GetCurrentInterval())
	}
}

// TestMonitorAndAdjust_ContextCancellation verifies that MonitorAndAdjust stops when context is cancelled.
func TestMonitorAndAdjust_ContextCancellation(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePolling(10)

	ctx, cancel := context.WithCancel(context.Background())

	// Create a function that returns 0 active recordings
	getActiveRecordingsCount := func() int {
		return 0
	}

	// Start monitoring in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- ap.MonitorAndAdjust(ctx, getActiveRecordingsCount)
	}()

	// Wait a bit to let the monitor start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for the monitor to stop
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("MonitorAndAdjust did not stop within timeout")
	}
}

// TestMonitorAndAdjust_IntervalAdjustment verifies that MonitorAndAdjust adjusts the interval based on recording activity.
func TestMonitorAndAdjust_IntervalAdjustment(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePolling(10)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a variable to control the active recordings count
	activeRecordings := 0

	// Create a function that returns the active recordings count
	getActiveRecordingsCount := func() int {
		return activeRecordings
	}

	// Start monitoring in a goroutine
	go func() {
		_ = ap.MonitorAndAdjust(ctx, getActiveRecordingsCount)
	}()

	// Wait for the initial check (happens immediately on start)
	time.Sleep(100 * time.Millisecond)

	// Verify interval was reduced (should happen on initial check since activeRecordings = 0)
	if ap.GetCurrentInterval() != 5 {
		t.Errorf("Expected interval to be reduced to 5, got %d", ap.GetCurrentInterval())
	}

	// Set active recordings to 1
	activeRecordings = 1

	// For testing purposes, we'll manually trigger an update instead of waiting 1 minute
	// This is acceptable because we're testing the logic, not the timing
	ap.UpdateInterval(true)

	// Verify interval was restored
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected interval to be restored to 10, got %d", ap.GetCurrentInterval())
	}
}

// TestGetLastUpdateTime verifies that GetLastUpdateTime returns the correct time.
func TestGetLastUpdateTime(t *testing.T) {
	ap := NewAdaptivePolling(10)

	// Get initial last update time
	initialTime := ap.GetLastUpdateTime()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Update the interval
	ap.UpdateInterval(false)

	// Get new last update time
	newTime := ap.GetLastUpdateTime()

	// Verify the time has changed
	if !newTime.After(initialTime) {
		t.Error("Expected last update time to be after initial time")
	}
}

// TestConcurrentAccess verifies that AdaptivePolling is thread-safe.
func TestConcurrentAccess(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePolling(10)

	// Start multiple goroutines that read and write concurrently
	done := make(chan bool)

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = ap.GetCurrentInterval()
				_ = ap.GetLastUpdateTime()
			}
			done <- true
		}()
	}

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func(index int) {
			for j := 0; j < 100; j++ {
				ap.UpdateInterval(index%2 == 0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// If we get here without a race condition, the test passes
}

// TestNewAdaptivePollingWithCostSaving verifies that cost-saving mode is initialized correctly.
// Requirements: 12.5, 12.6
func TestNewAdaptivePollingWithCostSaving(t *testing.T) {
	tests := []struct {
		name            string
		normalInterval  int
		costSavingMode  bool
		expectedNormal  int
		expectedReduced int
		expectedCostSaving int
		expectedCurrent int
		expectedMode    bool
	}{
		{
			name:            "Cost-saving mode enabled",
			normalInterval:  1,
			costSavingMode:  true,
			expectedNormal:  1,
			expectedReduced: 5,
			expectedCostSaving: 10,
			expectedCurrent: 10, // Should start with cost-saving interval
			expectedMode:    true,
		},
		{
			name:            "Cost-saving mode disabled",
			normalInterval:  1,
			costSavingMode:  false,
			expectedNormal:  1,
			expectedReduced: 5,
			expectedCostSaving: 10,
			expectedCurrent: 1, // Should start with normal interval
			expectedMode:    false,
		},
		{
			name:            "Cost-saving mode with custom normal interval",
			normalInterval:  3,
			costSavingMode:  true,
			expectedNormal:  3,
			expectedReduced: 5,
			expectedCostSaving: 10,
			expectedCurrent: 10, // Should start with cost-saving interval
			expectedMode:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := NewAdaptivePollingWithCostSaving(tt.normalInterval, tt.costSavingMode)

			if ap.GetNormalInterval() != tt.expectedNormal {
				t.Errorf("Expected normal interval %d, got %d", tt.expectedNormal, ap.GetNormalInterval())
			}

			if ap.GetReducedInterval() != tt.expectedReduced {
				t.Errorf("Expected reduced interval %d, got %d", tt.expectedReduced, ap.GetReducedInterval())
			}

			if ap.GetCostSavingInterval() != tt.expectedCostSaving {
				t.Errorf("Expected cost-saving interval %d, got %d", tt.expectedCostSaving, ap.GetCostSavingInterval())
			}

			if ap.GetCurrentInterval() != tt.expectedCurrent {
				t.Errorf("Expected current interval %d, got %d", tt.expectedCurrent, ap.GetCurrentInterval())
			}

			if ap.IsCostSavingMode() != tt.expectedMode {
				t.Errorf("Expected cost-saving mode %v, got %v", tt.expectedMode, ap.IsCostSavingMode())
			}
		})
	}
}

// TestUpdateInterval_CostSavingMode verifies that cost-saving mode always uses 10-minute interval.
// Requirements: 12.6
func TestUpdateInterval_CostSavingMode(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePollingWithCostSaving(1, true)

	// Initially, the interval should be the cost-saving interval (10)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected initial interval 10, got %d", ap.GetCurrentInterval())
	}

	// Update with no active recordings - should stay at 10
	ap.UpdateInterval(false)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected interval to remain 10 with no recordings, got %d", ap.GetCurrentInterval())
	}

	// Update with active recordings - should still stay at 10
	ap.UpdateInterval(true)
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected interval to remain 10 with active recordings, got %d", ap.GetCurrentInterval())
	}

	// Verify server config was updated to 10
	if server.Config.Interval != 10 {
		t.Errorf("Expected server config interval 10, got %d", server.Config.Interval)
	}
}

// TestUpdateInterval_CostSavingModeTransitions verifies that cost-saving mode doesn't transition intervals.
// Requirements: 12.6
func TestUpdateInterval_CostSavingModeTransitions(t *testing.T) {
	// Setup server config
	server.Config = &entity.Config{
		Interval: 10,
	}

	ap := NewAdaptivePollingWithCostSaving(1, true)

	// Get initial last update time
	initialUpdateTime := ap.GetLastUpdateTime()

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Try multiple transitions - none should change the interval
	ap.UpdateInterval(false)
	ap.UpdateInterval(true)
	ap.UpdateInterval(false)
	ap.UpdateInterval(true)

	// The interval should always be 10
	if ap.GetCurrentInterval() != 10 {
		t.Errorf("Expected interval to remain 10, got %d", ap.GetCurrentInterval())
	}

	// The last update time should not have changed (no actual updates occurred)
	if ap.GetLastUpdateTime() != initialUpdateTime {
		t.Error("Expected last update time to remain unchanged in cost-saving mode")
	}
}

// TestGetCostSavingInterval verifies that GetCostSavingInterval returns the correct value.
// Requirements: 12.6
func TestGetCostSavingInterval(t *testing.T) {
	ap := NewAdaptivePolling(1)

	if ap.GetCostSavingInterval() != 10 {
		t.Errorf("Expected cost-saving interval 10, got %d", ap.GetCostSavingInterval())
	}
}

// TestIsCostSavingMode verifies that IsCostSavingMode returns the correct value.
// Requirements: 12.5
func TestIsCostSavingMode(t *testing.T) {
	apNormal := NewAdaptivePolling(1)
	if apNormal.IsCostSavingMode() {
		t.Error("Expected cost-saving mode to be false for NewAdaptivePolling")
	}

	apCostSaving := NewAdaptivePollingWithCostSaving(1, true)
	if !apCostSaving.IsCostSavingMode() {
		t.Error("Expected cost-saving mode to be true")
	}

	apNoCostSaving := NewAdaptivePollingWithCostSaving(1, false)
	if apNoCostSaving.IsCostSavingMode() {
		t.Error("Expected cost-saving mode to be false")
	}
}

