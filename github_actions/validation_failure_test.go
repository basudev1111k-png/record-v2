package github_actions

import (
	"os"
	"testing"
)

// TestValidationFailure_WorkflowInputs verifies that validation failures are properly reported
// This test ensures that when validation fails, descriptive error messages are logged
// and the validation result indicates failure.
//
// Requirements: 5.10
func TestValidationFailure_WorkflowInputs(t *testing.T) {
	// Clear all environment variables to trigger validation failures
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	channels := []string{} // Empty channels list
	matrixJobCount := 0    // Invalid matrix job count

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	// Verify validation failed
	if result.Valid {
		t.Error("Expected validation to fail, but it passed")
	}

	// Verify errors are present
	if len(result.Errors) == 0 {
		t.Error("Expected validation errors, but got none")
	}

	// Verify specific error messages are descriptive
	expectedErrors := map[string]bool{
		"channels list cannot be empty":                                  false,
		"matrix_job_count must be at least 1, got 0":                     false,
		"GOFILE_API_KEY environment variable is required":                false,
		"FILESTER_API_KEY environment variable is required":              false,
	}

	for _, err := range result.Errors {
		if _, exists := expectedErrors[err]; exists {
			expectedErrors[err] = true
		}
	}

	// Verify all expected errors were found
	for expectedErr, found := range expectedErrors {
		if !found {
			t.Errorf("Expected error message not found: %s", expectedErr)
		}
	}
}

// TestValidationFailure_SetupMode verifies that setup validation failures are properly reported
// This test ensures that when setup validation fails, all errors are collected and reported.
//
// Requirements: 5.10
func TestValidationFailure_SetupMode(t *testing.T) {
	// Clear all environment variables to trigger validation failures
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	channels := []string{} // Empty channels list
	matrixJobCount := 25   // Invalid: exceeds limit

	result := validator.ValidateSetup(channels, matrixJobCount)

	// Verify validation failed
	if result.Valid {
		t.Error("Expected validation to fail, but it passed")
	}

	// Verify multiple errors are collected
	if len(result.Errors) < 6 {
		t.Errorf("Expected at least 6 errors, but got %d: %v", len(result.Errors), result.Errors)
	}

	// Verify errors contain descriptive messages
	errorMessages := make(map[string]bool)
	for _, err := range result.Errors {
		errorMessages[err] = true
	}

	// Check for key error messages (with descriptions in parentheses)
	requiredErrors := []string{
		"channels list cannot be empty",
		"GITHUB_TOKEN environment variable is required (GitHub API authentication token)",
		"GITHUB_REPOSITORY environment variable is required (GitHub repository identifier)",
		"GOFILE_API_KEY environment variable is required (Gofile upload API key)",
		"FILESTER_API_KEY environment variable is required (Filester upload API key)",
	}

	for _, requiredErr := range requiredErrors {
		found := false
		for actualErr := range errorMessages {
			if actualErr == requiredErr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error message not found: %s", requiredErr)
		}
	}
}

// TestValidationFailure_DescriptiveMessages verifies that error messages are descriptive
// This test ensures that validation error messages provide enough context to understand
// what went wrong and how to fix it.
//
// Requirements: 5.10
func TestValidationFailure_DescriptiveMessages(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name           string
		channels       []string
		matrixJobCount int
		setupEnv       func()
		cleanupEnv     func()
		expectedErrors []string
	}{
		{
			name:           "Empty channels",
			channels:       []string{},
			matrixJobCount: 5,
			setupEnv: func() {
				os.Setenv("GOFILE_API_KEY", "test-key")
				os.Setenv("FILESTER_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			expectedErrors: []string{
				"channels list cannot be empty",
			},
		},
		{
			name:           "Matrix job count too low",
			channels:       []string{"channel1"},
			matrixJobCount: 0,
			setupEnv: func() {
				os.Setenv("GOFILE_API_KEY", "test-key")
				os.Setenv("FILESTER_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			expectedErrors: []string{
				"matrix_job_count must be at least 1, got 0",
			},
		},
		{
			name:           "Matrix job count too high",
			channels:       []string{"channel1"},
			matrixJobCount: 25,
			setupEnv: func() {
				os.Setenv("GOFILE_API_KEY", "test-key")
				os.Setenv("FILESTER_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			expectedErrors: []string{
				"matrix_job_count cannot exceed 20 (GitHub Actions limit), got 25",
			},
		},
		{
			name:           "Missing API keys",
			channels:       []string{"channel1"},
			matrixJobCount: 5,
			setupEnv: func() {
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			cleanupEnv: func() {},
			expectedErrors: []string{
				"GOFILE_API_KEY environment variable is required",
				"FILESTER_API_KEY environment variable is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			result := validator.ValidateWorkflowInputs(tt.channels, tt.matrixJobCount)

			// Verify validation failed
			if result.Valid {
				t.Error("Expected validation to fail, but it passed")
			}

			// Verify expected errors are present
			for _, expectedErr := range tt.expectedErrors {
				found := false
				for _, actualErr := range result.Errors {
					if actualErr == expectedErr {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message not found: %s\nActual errors: %v", expectedErr, result.Errors)
				}
			}
		})
	}
}

// TestValidationFailure_ErrorLogging verifies that validation errors are properly structured
// This test ensures that the ValidationResult structure correctly tracks validation state
// and accumulates all errors.
//
// Requirements: 5.10
func TestValidationFailure_ErrorLogging(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Initially valid
	if !result.Valid {
		t.Error("Expected initial state to be valid")
	}

	// Add first error
	result.AddError("First validation error")

	// Should now be invalid
	if result.Valid {
		t.Error("Expected Valid to be false after adding error")
	}

	// Should have one error
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, but got %d", len(result.Errors))
	}

	// Add second error
	result.AddError("Second validation error")

	// Should still be invalid
	if result.Valid {
		t.Error("Expected Valid to remain false after adding second error")
	}

	// Should have two errors
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, but got %d", len(result.Errors))
	}

	// Verify error messages are preserved in order
	if result.Errors[0] != "First validation error" {
		t.Errorf("Expected first error to be 'First validation error', but got '%s'", result.Errors[0])
	}
	if result.Errors[1] != "Second validation error" {
		t.Errorf("Expected second error to be 'Second validation error', but got '%s'", result.Errors[1])
	}
}

// TestValidationFailure_PartialNtfyConfig verifies descriptive errors for partial ntfy configuration
// This test ensures that when ntfy is partially configured, the error messages clearly indicate
// which configuration values are missing.
//
// Requirements: 5.10
func TestValidationFailure_PartialNtfyConfig(t *testing.T) {
	// Set up all required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{"channel1"}
	matrixJobCount := 5

	tests := []struct {
		name          string
		setupEnv      func()
		cleanupEnv    func()
		expectedError string
	}{
		{
			name: "Server without topic",
			setupEnv: func() {
				os.Setenv("NTFY_SERVER_URL", "https://ntfy.sh")
				os.Unsetenv("NTFY_TOPIC")
			},
			cleanupEnv: func() {
				os.Unsetenv("NTFY_SERVER_URL")
			},
			expectedError: "NTFY_SERVER_URL is set but NTFY_TOPIC is missing",
		},
		{
			name: "Topic without server",
			setupEnv: func() {
				os.Unsetenv("NTFY_SERVER_URL")
				os.Setenv("NTFY_TOPIC", "my-topic")
			},
			cleanupEnv: func() {
				os.Unsetenv("NTFY_TOPIC")
			},
			expectedError: "NTFY_TOPIC is set but NTFY_SERVER_URL is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			result := validator.ValidateSetup(channels, matrixJobCount)

			// Verify validation failed
			if result.Valid {
				t.Error("Expected validation to fail with partial ntfy config, but it passed")
			}

			// Verify expected error is present
			found := false
			for _, err := range result.Errors {
				if err == tt.expectedError {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message not found: %s\nActual errors: %v", tt.expectedError, result.Errors)
			}
		})
	}
}
