package github_actions

import (
	"testing"

	"github.com/HeapOfChaos/goondvr/entity"
)

func TestNewQualitySelector(t *testing.T) {
	qs := NewQualitySelector()

	if qs == nil {
		t.Fatal("NewQualitySelector returned nil")
	}

	if qs.GetPreferredResolution() != 2160 {
		t.Errorf("Expected preferred resolution 2160, got %d", qs.GetPreferredResolution())
	}

	if qs.GetPreferredFramerate() != 60 {
		t.Errorf("Expected preferred framerate 60, got %d", qs.GetPreferredFramerate())
	}
}

func TestSelectQuality_Priority1_2160p60(t *testing.T) {
	qs := NewQualitySelector()

	availableQualities := []Quality{
		{Resolution: 720, Framerate: 30},
		{Resolution: 1080, Framerate: 60},
		{Resolution: 2160, Framerate: 60}, // Should select this
		{Resolution: 1080, Framerate: 30},
	}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 2160 {
		t.Errorf("Expected resolution 2160, got %d", result.Resolution)
	}
	if result.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", result.Framerate)
	}
	if result.Actual != "2160p60" {
		t.Errorf("Expected actual quality '2160p60', got '%s'", result.Actual)
	}
}

func TestSelectQuality_Priority2_1080p60(t *testing.T) {
	qs := NewQualitySelector()

	availableQualities := []Quality{
		{Resolution: 720, Framerate: 30},
		{Resolution: 1080, Framerate: 60}, // Should select this
		{Resolution: 1080, Framerate: 30},
		{Resolution: 720, Framerate: 60},
	}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 1080 {
		t.Errorf("Expected resolution 1080, got %d", result.Resolution)
	}
	if result.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", result.Framerate)
	}
	if result.Actual != "1080p60" {
		t.Errorf("Expected actual quality '1080p60', got '%s'", result.Actual)
	}
}

func TestSelectQuality_Priority3_720p60(t *testing.T) {
	qs := NewQualitySelector()

	availableQualities := []Quality{
		{Resolution: 720, Framerate: 30},
		{Resolution: 720, Framerate: 60}, // Should select this
		{Resolution: 480, Framerate: 60},
		{Resolution: 1080, Framerate: 30},
	}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 720 {
		t.Errorf("Expected resolution 720, got %d", result.Resolution)
	}
	if result.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", result.Framerate)
	}
	if result.Actual != "720p60" {
		t.Errorf("Expected actual quality '720p60', got '%s'", result.Actual)
	}
}

func TestSelectQuality_Priority4_HighestAvailable_ByResolution(t *testing.T) {
	qs := NewQualitySelector()

	availableQualities := []Quality{
		{Resolution: 720, Framerate: 30},
		{Resolution: 1080, Framerate: 30}, // Should select this (highest resolution)
		{Resolution: 480, Framerate: 30},
	}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 1080 {
		t.Errorf("Expected resolution 1080, got %d", result.Resolution)
	}
	if result.Framerate != 30 {
		t.Errorf("Expected framerate 30, got %d", result.Framerate)
	}
	if result.Actual != "1080p30" {
		t.Errorf("Expected actual quality '1080p30', got '%s'", result.Actual)
	}
}

func TestSelectQuality_Priority4_HighestAvailable_ByFramerate(t *testing.T) {
	qs := NewQualitySelector()

	// No 60fps options at priority resolutions, so should fall to Priority 4
	availableQualities := []Quality{
		{Resolution: 1080, Framerate: 30},
		{Resolution: 1080, Framerate: 50}, // Should select this (same resolution, higher framerate)
		{Resolution: 480, Framerate: 30},
	}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 1080 {
		t.Errorf("Expected resolution 1080, got %d", result.Resolution)
	}
	if result.Framerate != 50 {
		t.Errorf("Expected framerate 50, got %d", result.Framerate)
	}
	if result.Actual != "1080p50" {
		t.Errorf("Expected actual quality '1080p50', got '%s'", result.Actual)
	}
}

func TestSelectQuality_EmptyList_ReturnsDefault(t *testing.T) {
	qs := NewQualitySelector()

	availableQualities := []Quality{}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 720 {
		t.Errorf("Expected default resolution 720, got %d", result.Resolution)
	}
	if result.Framerate != 30 {
		t.Errorf("Expected default framerate 30, got %d", result.Framerate)
	}
	if result.Actual != "720p30" {
		t.Errorf("Expected default quality '720p30', got '%s'", result.Actual)
	}
}

func TestSelectQuality_SingleQuality(t *testing.T) {
	qs := NewQualitySelector()

	availableQualities := []Quality{
		{Resolution: 480, Framerate: 25},
	}

	result := qs.SelectQuality(availableQualities)

	if result.Resolution != 480 {
		t.Errorf("Expected resolution 480, got %d", result.Resolution)
	}
	if result.Framerate != 25 {
		t.Errorf("Expected framerate 25, got %d", result.Framerate)
	}
	if result.Actual != "480p25" {
		t.Errorf("Expected actual quality '480p25', got '%s'", result.Actual)
	}
}

func TestSelectQuality_FallbackChain(t *testing.T) {
	tests := []struct {
		name               string
		availableQualities []Quality
		expectedResolution int
		expectedFramerate  int
		expectedActual     string
	}{
		{
			name: "Priority 1: 2160p60 available",
			availableQualities: []Quality{
				{Resolution: 720, Framerate: 30},
				{Resolution: 2160, Framerate: 60},
			},
			expectedResolution: 2160,
			expectedFramerate:  60,
			expectedActual:     "2160p60",
		},
		{
			name: "Priority 2: 1080p60 available (no 2160p60)",
			availableQualities: []Quality{
				{Resolution: 720, Framerate: 30},
				{Resolution: 1080, Framerate: 60},
				{Resolution: 2160, Framerate: 30},
			},
			expectedResolution: 1080,
			expectedFramerate:  60,
			expectedActual:     "1080p60",
		},
		{
			name: "Priority 3: 720p60 available (no 2160p60 or 1080p60)",
			availableQualities: []Quality{
				{Resolution: 720, Framerate: 30},
				{Resolution: 720, Framerate: 60},
				{Resolution: 480, Framerate: 60},
			},
			expectedResolution: 720,
			expectedFramerate:  60,
			expectedActual:     "720p60",
		},
		{
			name: "Priority 4: Highest available (no 60fps options)",
			availableQualities: []Quality{
				{Resolution: 720, Framerate: 30},
				{Resolution: 1080, Framerate: 30},
				{Resolution: 480, Framerate: 30},
			},
			expectedResolution: 1080,
			expectedFramerate:  30,
			expectedActual:     "1080p30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := NewQualitySelector()
			result := qs.SelectQuality(tt.availableQualities)

			if result.Resolution != tt.expectedResolution {
				t.Errorf("Expected resolution %d, got %d", tt.expectedResolution, result.Resolution)
			}
			if result.Framerate != tt.expectedFramerate {
				t.Errorf("Expected framerate %d, got %d", tt.expectedFramerate, result.Framerate)
			}
			if result.Actual != tt.expectedActual {
				t.Errorf("Expected actual quality '%s', got '%s'", tt.expectedActual, result.Actual)
			}
		})
	}
}

func TestApplyQualitySettings(t *testing.T) {
	qs := NewQualitySelector()

	// Create a channel config with initial settings
	config := &entity.ChannelConfig{
		Username:   "testuser",
		Site:       "chaturbate",
		Resolution: 720,
		Framerate:  30,
	}

	// Create quality settings to apply
	settings := QualitySettings{
		Resolution: 2160,
		Framerate:  60,
		Actual:     "2160p60",
	}

	// Apply the quality settings
	qs.ApplyQualitySettings(config, settings)

	// Verify the config was updated
	if config.Resolution != 2160 {
		t.Errorf("Expected resolution 2160, got %d", config.Resolution)
	}
	if config.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", config.Framerate)
	}

	// Verify other fields were not modified
	if config.Username != "testuser" {
		t.Errorf("Username should not be modified, got %s", config.Username)
	}
	if config.Site != "chaturbate" {
		t.Errorf("Site should not be modified, got %s", config.Site)
	}
}

func TestApplyQualitySettings_OverridesExisting(t *testing.T) {
	qs := NewQualitySelector()

	// Create a channel config with existing quality settings
	config := &entity.ChannelConfig{
		Username:   "testuser",
		Resolution: 1080,
		Framerate:  30,
	}

	// Apply new quality settings
	settings := QualitySettings{
		Resolution: 720,
		Framerate:  60,
		Actual:     "720p60",
	}

	qs.ApplyQualitySettings(config, settings)

	// Verify the config was overridden
	if config.Resolution != 720 {
		t.Errorf("Expected resolution 720 (overridden), got %d", config.Resolution)
	}
	if config.Framerate != 60 {
		t.Errorf("Expected framerate 60 (overridden), got %d", config.Framerate)
	}
}

func TestApplyQualitySettings_LogsQuality(t *testing.T) {
	qs := NewQualitySelector()

	// Create a channel config
	config := &entity.ChannelConfig{
		Username:   "testuser",
		Site:       "chaturbate",
		Resolution: 720,
		Framerate:  30,
	}

	// Apply quality settings
	settings := QualitySettings{
		Resolution: 2160,
		Framerate:  60,
		Actual:     "2160p60",
	}

	// Note: This test verifies that ApplyQualitySettings executes without error
	// and properly sets the configuration. The actual log output is verified
	// through manual testing or integration tests.
	qs.ApplyQualitySettings(config, settings)

	// Verify the config was updated
	if config.Resolution != 2160 {
		t.Errorf("Expected resolution 2160, got %d", config.Resolution)
	}
	if config.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", config.Framerate)
	}
}

func TestQualitySettings_ActualFormat(t *testing.T) {
	tests := []struct {
		resolution int
		framerate  int
		expected   string
	}{
		{2160, 60, "2160p60"},
		{1080, 60, "1080p60"},
		{720, 60, "720p60"},
		{1080, 30, "1080p30"},
		{720, 30, "720p30"},
		{480, 25, "480p25"},
	}

	qs := NewQualitySelector()

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			availableQualities := []Quality{
				{Resolution: tt.resolution, Framerate: tt.framerate},
			}

			result := qs.SelectQuality(availableQualities)

			if result.Actual != tt.expected {
				t.Errorf("Expected actual quality '%s', got '%s'", tt.expected, result.Actual)
			}
		})
	}
}

func TestSelectQuality_ComplexScenario(t *testing.T) {
	qs := NewQualitySelector()

	// Simulate a realistic scenario with multiple quality options
	availableQualities := []Quality{
		{Resolution: 480, Framerate: 30},
		{Resolution: 720, Framerate: 30},
		{Resolution: 720, Framerate: 60},
		{Resolution: 1080, Framerate: 30},
		{Resolution: 1080, Framerate: 60},
		{Resolution: 2160, Framerate: 30},
		// Note: 2160p60 is NOT available
	}

	result := qs.SelectQuality(availableQualities)

	// Should select 1080p60 (Priority 2) since 2160p60 is not available
	if result.Resolution != 1080 {
		t.Errorf("Expected resolution 1080, got %d", result.Resolution)
	}
	if result.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", result.Framerate)
	}
	if result.Actual != "1080p60" {
		t.Errorf("Expected actual quality '1080p60', got '%s'", result.Actual)
	}
}

func TestSelectQuality_PreferResolutionOverFramerate(t *testing.T) {
	qs := NewQualitySelector()

	// When no 60fps options match priorities, should prefer higher resolution
	availableQualities := []Quality{
		{Resolution: 720, Framerate: 60},
		{Resolution: 2160, Framerate: 30}, // Higher resolution but lower framerate
		{Resolution: 1080, Framerate: 30},
	}

	result := qs.SelectQuality(availableQualities)

	// Should select 720p60 (Priority 3) over 2160p30
	if result.Resolution != 720 {
		t.Errorf("Expected resolution 720, got %d", result.Resolution)
	}
	if result.Framerate != 60 {
		t.Errorf("Expected framerate 60, got %d", result.Framerate)
	}
	if result.Actual != "720p60" {
		t.Errorf("Expected actual quality '720p60', got '%s'", result.Actual)
	}
}

func TestDetectAvailableQualities_ValidURL(t *testing.T) {
	qs := NewQualitySelector()

	// Test with a valid stream URL (placeholder implementation)
	streamURL := "https://example.com/stream/master.m3u8"

	qualities, err := qs.DetectAvailableQualities(streamURL)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if qualities == nil {
		t.Fatal("Expected qualities slice, got nil")
	}

	if len(qualities) == 0 {
		t.Error("Expected at least one quality option")
	}

	// Verify that qualities contain expected resolutions
	hasHighQuality := false
	for _, q := range qualities {
		if q.Resolution >= 1080 && q.Framerate >= 30 {
			hasHighQuality = true
			break
		}
	}

	if !hasHighQuality {
		t.Error("Expected at least one high quality option (1080p or better)")
	}
}

func TestDetectAvailableQualities_EmptyURL(t *testing.T) {
	qs := NewQualitySelector()

	// Test with empty stream URL
	streamURL := ""

	qualities, err := qs.DetectAvailableQualities(streamURL)

	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}

	if qualities != nil {
		t.Errorf("Expected nil qualities for empty URL, got %v", qualities)
	}
}

func TestDetectAvailableQualities_ReturnsValidQualities(t *testing.T) {
	qs := NewQualitySelector()

	streamURL := "https://example.com/stream/master.m3u8"

	qualities, err := qs.DetectAvailableQualities(streamURL)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all returned qualities have valid resolution and framerate
	for i, q := range qualities {
		if q.Resolution <= 0 {
			t.Errorf("Quality %d has invalid resolution: %d", i, q.Resolution)
		}
		if q.Framerate <= 0 {
			t.Errorf("Quality %d has invalid framerate: %d", i, q.Framerate)
		}
	}
}

func TestDetectAvailableQualities_Integration(t *testing.T) {
	qs := NewQualitySelector()

	// Test the integration between DetectAvailableQualities and SelectQuality
	streamURL := "https://example.com/stream/master.m3u8"

	qualities, err := qs.DetectAvailableQualities(streamURL)
	if err != nil {
		t.Fatalf("Expected no error from DetectAvailableQualities, got %v", err)
	}

	// Use the detected qualities to select the best one
	selected := qs.SelectQuality(qualities)

	// Verify the selected quality is valid
	if selected.Resolution <= 0 {
		t.Errorf("Selected quality has invalid resolution: %d", selected.Resolution)
	}
	if selected.Framerate <= 0 {
		t.Errorf("Selected quality has invalid framerate: %d", selected.Framerate)
	}
	if selected.Actual == "" {
		t.Error("Selected quality has empty Actual field")
	}

	// Verify the selected quality is one of the detected qualities
	found := false
	for _, q := range qualities {
		if q.Resolution == selected.Resolution && q.Framerate == selected.Framerate {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Selected quality %s not found in detected qualities", selected.Actual)
	}
}
