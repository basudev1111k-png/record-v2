package github_actions

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewStreamFailureRecovery(t *testing.T) {
	tests := []struct {
		name          string
		retryInterval time.Duration
		wantInterval  time.Duration
	}{
		{
			name:          "with custom retry interval",
			retryInterval: 10 * time.Minute,
			wantInterval:  10 * time.Minute,
		},
		{
			name:          "with zero retry interval (should use default)",
			retryInterval: 0,
			wantInterval:  5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
			sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", tt.retryInterval)

			if sfr == nil {
				t.Fatal("NewStreamFailureRecovery returned nil")
			}

			if sfr.sessionID != "test-session" {
				t.Errorf("sessionID = %s, want test-session", sfr.sessionID)
			}

			if sfr.matrixJobID != "test-job" {
				t.Errorf("matrixJobID = %s, want test-job", sfr.matrixJobID)
			}

			if sfr.retryInterval != tt.wantInterval {
				t.Errorf("retryInterval = %s, want %s", sfr.retryInterval, tt.wantInterval)
			}

			if sfr.failureCount == nil {
				t.Error("failureCount map is nil")
			}

			if sfr.lastFailureTime == nil {
				t.Error("lastFailureTime map is nil")
			}
		})
	}
}

func TestLogStreamFailure(t *testing.T) {
	ctx := context.Background()
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	tests := []struct {
		name                string
		info                StreamFailureInfo
		expectedFailureCount int
		wantRetryInterval   time.Duration
	}{
		{
			name: "first failure",
			info: StreamFailureInfo{
				Channel:     "testchannel",
				Site:        "chaturbate",
				Error:       errors.New("channel offline"),
				FailureType: "offline",
				Timestamp:   time.Now(),
			},
			expectedFailureCount: 1,
			wantRetryInterval:   5 * time.Minute,
		},
		{
			name: "second failure",
			info: StreamFailureInfo{
				Channel:     "testchannel",
				Site:        "chaturbate",
				Error:       errors.New("channel offline"),
				FailureType: "offline",
				Timestamp:   time.Now(),
			},
			expectedFailureCount: 2,
			wantRetryInterval:   5 * time.Minute,
		},
		{
			name: "third failure (should trigger notification)",
			info: StreamFailureInfo{
				Channel:     "testchannel",
				Site:        "chaturbate",
				Error:       errors.New("channel offline"),
				FailureType: "offline",
				Timestamp:   time.Now(),
			},
			expectedFailureCount: 3,
			wantRetryInterval:   5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryInterval := sfr.LogStreamFailure(ctx, tt.info)

			if retryInterval != tt.wantRetryInterval {
				t.Errorf("LogStreamFailure() returned interval = %s, want %s", retryInterval, tt.wantRetryInterval)
			}

			failureCount := sfr.GetFailureCount(tt.info.Channel, tt.info.Site)
			if failureCount != tt.expectedFailureCount {
				t.Errorf("GetFailureCount() = %d, want %d", failureCount, tt.expectedFailureCount)
			}

			lastFailureTime := sfr.GetLastFailureTime(tt.info.Channel, tt.info.Site)
			if lastFailureTime.IsZero() {
				t.Error("GetLastFailureTime() returned zero time")
			}
		})
	}
}

func TestLogStreamRecovery(t *testing.T) {
	ctx := context.Background()
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	// Simulate multiple failures
	info := StreamFailureInfo{
		Channel:     "testchannel",
		Site:        "chaturbate",
		Error:       errors.New("channel offline"),
		FailureType: "offline",
		Timestamp:   time.Now(),
	}

	// Log 3 failures
	for i := 0; i < 3; i++ {
		sfr.LogStreamFailure(ctx, info)
	}

	// Verify failure count before recovery
	failureCount := sfr.GetFailureCount("testchannel", "chaturbate")
	if failureCount != 3 {
		t.Errorf("GetFailureCount() before recovery = %d, want 3", failureCount)
	}

	// Log recovery
	sfr.LogStreamRecovery(ctx, "testchannel", "chaturbate")

	// Verify failure count is reset after recovery
	failureCount = sfr.GetFailureCount("testchannel", "chaturbate")
	if failureCount != 0 {
		t.Errorf("GetFailureCount() after recovery = %d, want 0", failureCount)
	}

	// Verify last failure time is cleared
	lastFailureTime := sfr.GetLastFailureTime("testchannel", "chaturbate")
	if !lastFailureTime.IsZero() {
		t.Error("GetLastFailureTime() after recovery should be zero time")
	}
}

func TestGetFailureStatistics(t *testing.T) {
	ctx := context.Background()
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	// Log failures for multiple channels
	channels := []struct {
		channel string
		site    string
		count   int
	}{
		{"channel1", "chaturbate", 2},
		{"channel2", "stripchat", 3},
		{"channel3", "chaturbate", 1},
	}

	for _, ch := range channels {
		for i := 0; i < ch.count; i++ {
			info := StreamFailureInfo{
				Channel:     ch.channel,
				Site:        ch.site,
				Error:       errors.New("test error"),
				FailureType: "offline",
				Timestamp:   time.Now(),
			}
			sfr.LogStreamFailure(ctx, info)
		}
	}

	// Get statistics
	stats := sfr.GetFailureStatistics()

	// Verify statistics for each channel
	for _, ch := range channels {
		channelKey := ch.site + "/" + ch.channel
		stat, exists := stats[channelKey]
		if !exists {
			t.Errorf("Statistics not found for channel %s", channelKey)
			continue
		}

		if stat.FailureCount != ch.count {
			t.Errorf("FailureCount for %s = %d, want %d", channelKey, stat.FailureCount, ch.count)
		}

		if stat.LastFailureTime.IsZero() {
			t.Errorf("LastFailureTime for %s is zero", channelKey)
		}
	}
}

func TestSetRetryInterval(t *testing.T) {
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	tests := []struct {
		name         string
		newInterval  time.Duration
		wantInterval time.Duration
	}{
		{
			name:         "update to 10 minutes",
			newInterval:  10 * time.Minute,
			wantInterval: 10 * time.Minute,
		},
		{
			name:         "update to 1 minute",
			newInterval:  1 * time.Minute,
			wantInterval: 1 * time.Minute,
		},
		{
			name:         "zero interval (should not update)",
			newInterval:  0,
			wantInterval: 1 * time.Minute, // Should keep previous value
		},
		{
			name:         "negative interval (should not update)",
			newInterval:  -5 * time.Minute,
			wantInterval: 1 * time.Minute, // Should keep previous value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfr.SetRetryInterval(tt.newInterval)
			gotInterval := sfr.GetRetryInterval()
			if gotInterval != tt.wantInterval {
				t.Errorf("GetRetryInterval() = %s, want %s", gotInterval, tt.wantInterval)
			}
		})
	}
}

func TestResetFailureCount(t *testing.T) {
	ctx := context.Background()
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	// Log failures for a channel
	info := StreamFailureInfo{
		Channel:     "testchannel",
		Site:        "chaturbate",
		Error:       errors.New("test error"),
		FailureType: "offline",
		Timestamp:   time.Now(),
	}

	for i := 0; i < 5; i++ {
		sfr.LogStreamFailure(ctx, info)
	}

	// Verify failure count before reset
	failureCount := sfr.GetFailureCount("testchannel", "chaturbate")
	if failureCount != 5 {
		t.Errorf("GetFailureCount() before reset = %d, want 5", failureCount)
	}

	// Reset failure count
	sfr.ResetFailureCount("testchannel", "chaturbate")

	// Verify failure count after reset
	failureCount = sfr.GetFailureCount("testchannel", "chaturbate")
	if failureCount != 0 {
		t.Errorf("GetFailureCount() after reset = %d, want 0", failureCount)
	}

	// Verify last failure time is cleared
	lastFailureTime := sfr.GetLastFailureTime("testchannel", "chaturbate")
	if !lastFailureTime.IsZero() {
		t.Error("GetLastFailureTime() after reset should be zero time")
	}
}

func TestResetAllFailureCounts(t *testing.T) {
	ctx := context.Background()
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	// Log failures for multiple channels
	channels := []struct {
		channel string
		site    string
	}{
		{"channel1", "chaturbate"},
		{"channel2", "stripchat"},
		{"channel3", "chaturbate"},
	}

	for _, ch := range channels {
		info := StreamFailureInfo{
			Channel:     ch.channel,
			Site:        ch.site,
			Error:       errors.New("test error"),
			FailureType: "offline",
			Timestamp:   time.Now(),
		}
		sfr.LogStreamFailure(ctx, info)
	}

	// Verify failures are recorded
	stats := sfr.GetFailureStatistics()
	if len(stats) != len(channels) {
		t.Errorf("GetFailureStatistics() returned %d entries, want %d", len(stats), len(channels))
	}

	// Reset all failure counts
	sfr.ResetAllFailureCounts()

	// Verify all failure counts are cleared
	stats = sfr.GetFailureStatistics()
	if len(stats) != 0 {
		t.Errorf("GetFailureStatistics() after reset returned %d entries, want 0", len(stats))
	}

	// Verify individual failure counts are zero
	for _, ch := range channels {
		failureCount := sfr.GetFailureCount(ch.channel, ch.site)
		if failureCount != 0 {
			t.Errorf("GetFailureCount(%s, %s) after reset = %d, want 0", ch.channel, ch.site, failureCount)
		}
	}
}

func TestMultipleChannelsIsolation(t *testing.T) {
	ctx := context.Background()
	healthMonitor := NewHealthMonitor("test-status.json", []Notifier{})
	sfr := NewStreamFailureRecovery(healthMonitor, "test-session", "test-job", 5*time.Minute)

	// Log failures for channel1
	info1 := StreamFailureInfo{
		Channel:     "channel1",
		Site:        "chaturbate",
		Error:       errors.New("test error"),
		FailureType: "offline",
		Timestamp:   time.Now(),
	}
	sfr.LogStreamFailure(ctx, info1)
	sfr.LogStreamFailure(ctx, info1)

	// Log failures for channel2
	info2 := StreamFailureInfo{
		Channel:     "channel2",
		Site:        "stripchat",
		Error:       errors.New("test error"),
		FailureType: "network",
		Timestamp:   time.Now(),
	}
	sfr.LogStreamFailure(ctx, info2)
	sfr.LogStreamFailure(ctx, info2)
	sfr.LogStreamFailure(ctx, info2)

	// Verify failure counts are independent
	count1 := sfr.GetFailureCount("channel1", "chaturbate")
	if count1 != 2 {
		t.Errorf("GetFailureCount(channel1) = %d, want 2", count1)
	}

	count2 := sfr.GetFailureCount("channel2", "stripchat")
	if count2 != 3 {
		t.Errorf("GetFailureCount(channel2) = %d, want 3", count2)
	}

	// Reset channel1 and verify channel2 is not affected
	sfr.ResetFailureCount("channel1", "chaturbate")

	count1 = sfr.GetFailureCount("channel1", "chaturbate")
	if count1 != 0 {
		t.Errorf("GetFailureCount(channel1) after reset = %d, want 0", count1)
	}

	count2 = sfr.GetFailureCount("channel2", "stripchat")
	if count2 != 3 {
		t.Errorf("GetFailureCount(channel2) after channel1 reset = %d, want 3 (should be unchanged)", count2)
	}
}
