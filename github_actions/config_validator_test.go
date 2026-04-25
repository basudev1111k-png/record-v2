package github_actions

import (
	"os"
	"testing"
)

// TestValidateWorkflowInputs_ValidConfiguration verifies validation passes with valid inputs
func TestValidateWorkflowInputs_ValidConfiguration(t *testing.T) {
	// Set up environment variables
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1", "stripchat:user2"}
	matrixJobCount := 5

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if !result.Valid {
		t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, but got %d errors", len(result.Errors))
	}
}

// TestValidateWorkflowInputs_EmptyChannels verifies validation fails with empty channels
func TestValidateWorkflowInputs_EmptyChannels(t *testing.T) {
	// Set up environment variables
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{} // Empty channels list
	matrixJobCount := 5

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with empty channels, but it passed")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one error for empty channels")
	}

	// Check that the error message mentions channels
	foundChannelError := false
	for _, err := range result.Errors {
		if err == "channels list cannot be empty" {
			foundChannelError = true
			break
		}
	}
	if !foundChannelError {
		t.Errorf("Expected error about empty channels, but got: %v", result.Errors)
	}
}

// TestValidateWorkflowInputs_MatrixJobCountTooLow verifies validation fails when matrix_job_count < 1
func TestValidateWorkflowInputs_MatrixJobCountTooLow(t *testing.T) {
	// Set up environment variables
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 0 // Invalid: too low

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with matrix_job_count = 0, but it passed")
	}

	// Check that the error message mentions matrix_job_count
	foundMatrixError := false
	for _, err := range result.Errors {
		if err == "matrix_job_count must be at least 1, got 0" {
			foundMatrixError = true
			break
		}
	}
	if !foundMatrixError {
		t.Errorf("Expected error about matrix_job_count being too low, but got: %v", result.Errors)
	}
}

// TestValidateWorkflowInputs_MatrixJobCountTooHigh verifies validation fails when matrix_job_count > 20
func TestValidateWorkflowInputs_MatrixJobCountTooHigh(t *testing.T) {
	// Set up environment variables
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 25 // Invalid: exceeds GitHub Actions limit

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with matrix_job_count = 25, but it passed")
	}

	// Check that the error message mentions the limit
	foundMatrixError := false
	for _, err := range result.Errors {
		if err == "matrix_job_count cannot exceed 20 (GitHub Actions limit), got 25" {
			foundMatrixError = true
			break
		}
	}
	if !foundMatrixError {
		t.Errorf("Expected error about matrix_job_count exceeding limit, but got: %v", result.Errors)
	}
}

// TestValidateWorkflowInputs_MatrixJobCountBoundaries verifies boundary values for matrix_job_count
func TestValidateWorkflowInputs_MatrixJobCountBoundaries(t *testing.T) {
	// Set up environment variables
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer func() {
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}

	tests := []struct {
		name           string
		matrixJobCount int
		shouldPass     bool
	}{
		{"minimum valid (1)", 1, true},
		{"maximum valid (20)", 20, true},
		{"below minimum (0)", 0, false},
		{"above maximum (21)", 21, false},
		{"negative value", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateWorkflowInputs(channels, tt.matrixJobCount)

			if tt.shouldPass && !result.Valid {
				t.Errorf("Expected validation to pass for matrix_job_count=%d, but got errors: %v",
					tt.matrixJobCount, result.Errors)
			}

			if !tt.shouldPass && result.Valid {
				t.Errorf("Expected validation to fail for matrix_job_count=%d, but it passed",
					tt.matrixJobCount)
			}
		})
	}
}

// TestValidateWorkflowInputs_MissingGofileAPIKey verifies validation fails without Gofile API key
func TestValidateWorkflowInputs_MissingGofileAPIKey(t *testing.T) {
	// Set up environment variables (missing GOFILE_API_KEY)
	os.Unsetenv("GOFILE_API_KEY")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	defer os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail without GOFILE_API_KEY, but it passed")
	}

	// Check that the error message mentions Gofile API key
	foundGofileError := false
	for _, err := range result.Errors {
		if err == "GOFILE_API_KEY environment variable is required" {
			foundGofileError = true
			break
		}
	}
	if !foundGofileError {
		t.Errorf("Expected error about missing GOFILE_API_KEY, but got: %v", result.Errors)
	}
}

// TestValidateWorkflowInputs_MissingFilesterAPIKey verifies validation fails without Filester API key
func TestValidateWorkflowInputs_MissingFilesterAPIKey(t *testing.T) {
	// Set up environment variables (missing FILESTER_API_KEY)
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Unsetenv("FILESTER_API_KEY")
	defer os.Unsetenv("GOFILE_API_KEY")

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail without FILESTER_API_KEY, but it passed")
	}

	// Check that the error message mentions Filester API key
	foundFilesterError := false
	for _, err := range result.Errors {
		if err == "FILESTER_API_KEY environment variable is required" {
			foundFilesterError = true
			break
		}
	}
	if !foundFilesterError {
		t.Errorf("Expected error about missing FILESTER_API_KEY, but got: %v", result.Errors)
	}
}

// TestValidateWorkflowInputs_MultipleErrors verifies all errors are collected
func TestValidateWorkflowInputs_MultipleErrors(t *testing.T) {
	// Set up environment with multiple issues
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	channels := []string{} // Empty channels
	matrixJobCount := 0    // Invalid matrix job count

	result := validator.ValidateWorkflowInputs(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with multiple errors, but it passed")
	}

	// Should have at least 4 errors: empty channels, matrix_job_count too low,
	// missing GOFILE_API_KEY, missing FILESTER_API_KEY
	if len(result.Errors) < 4 {
		t.Errorf("Expected at least 4 errors, but got %d: %v", len(result.Errors), result.Errors)
	}
}

// TestValidateEnvironmentVariables_AllPresent verifies validation passes with all env vars
func TestValidateEnvironmentVariables_AllPresent(t *testing.T) {
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
	result := validator.ValidateEnvironmentVariables()

	if !result.Valid {
		t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, but got %d errors", len(result.Errors))
	}
}

// TestValidateEnvironmentVariables_MissingVariables verifies validation fails with missing env vars
func TestValidateEnvironmentVariables_MissingVariables(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	result := validator.ValidateEnvironmentVariables()

	if result.Valid {
		t.Error("Expected validation to fail with missing env vars, but it passed")
	}

	// Should have 4 errors (one for each missing variable)
	if len(result.Errors) != 4 {
		t.Errorf("Expected 4 errors, but got %d: %v", len(result.Errors), result.Errors)
	}
}

// TestParseMatrixJobCount_ValidInput verifies parsing of valid integer strings
func TestParseMatrixJobCount_ValidInput(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		input    string
		expected int
	}{
		{"1", 1},
		{"5", 5},
		{"10", 10},
		{"20", 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := validator.ParseMatrixJobCount(tt.input)

			if err != nil {
				t.Errorf("Expected no error for input '%s', but got: %v", tt.input, err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d, but got %d", tt.expected, result)
			}
		})
	}
}

// TestParseMatrixJobCount_InvalidInput verifies parsing fails with invalid input
func TestParseMatrixJobCount_InvalidInput(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"non-numeric", "abc"},
		{"decimal", "5.5"},
		{"with spaces", " 5 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ParseMatrixJobCount(tt.input)

			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", tt.input)
			}
		})
	}
}

// TestValidatePollingInterval_ValidValues verifies validation passes with positive intervals
func TestValidatePollingInterval_ValidValues(t *testing.T) {
	validator := NewConfigValidator()

	tests := []int{1, 5, 10, 60}

	for _, interval := range tests {
		t.Run(string(rune(interval)), func(t *testing.T) {
			err := validator.ValidatePollingInterval(interval)

			if err != nil {
				t.Errorf("Expected no error for interval %d, but got: %v", interval, err)
			}
		})
	}
}

// TestValidatePollingInterval_InvalidValues verifies validation fails with non-positive intervals
func TestValidatePollingInterval_InvalidValues(t *testing.T) {
	validator := NewConfigValidator()

	tests := []int{0, -1, -10}

	for _, interval := range tests {
		t.Run(string(rune(interval)), func(t *testing.T) {
			err := validator.ValidatePollingInterval(interval)

			if err == nil {
				t.Errorf("Expected error for interval %d, but got none", interval)
			}
		})
	}
}

// TestNewConfigValidator verifies constructor creates a valid instance
func TestNewConfigValidator(t *testing.T) {
	validator := NewConfigValidator()

	if validator == nil {
		t.Error("Expected NewConfigValidator to return a non-nil instance")
	}
}

// TestValidationResult_AddError verifies error accumulation
func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Initially valid with no errors
	if !result.Valid {
		t.Error("Expected initial state to be valid")
	}
	if len(result.Errors) != 0 {
		t.Error("Expected no initial errors")
	}

	// Add first error
	result.AddError("first error")
	if result.Valid {
		t.Error("Expected Valid to be false after adding error")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, but got %d", len(result.Errors))
	}

	// Add second error
	result.AddError("second error")
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, but got %d", len(result.Errors))
	}

	// Verify error messages
	if result.Errors[0] != "first error" {
		t.Errorf("Expected first error to be 'first error', but got '%s'", result.Errors[0])
	}
	if result.Errors[1] != "second error" {
		t.Errorf("Expected second error to be 'second error', but got '%s'", result.Errors[1])
	}
}

// TestValidateSetup_AllValid verifies validation passes with all required configuration
func TestValidateSetup_AllValid(t *testing.T) {
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
	channels := []string{"chaturbate:user1", "stripchat:user2"}
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	if !result.Valid {
		t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, but got %d errors", len(result.Errors))
	}
}

// TestValidateSetup_MissingAllSecrets verifies validation fails with all secrets missing
func TestValidateSetup_MissingAllSecrets(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with missing secrets, but it passed")
	}

	// Should have at least 4 errors (one for each missing secret)
	if len(result.Errors) < 4 {
		t.Errorf("Expected at least 4 errors, but got %d: %v", len(result.Errors), result.Errors)
	}
}

// TestValidateSetup_EmptyChannels verifies validation fails with empty channels
func TestValidateSetup_EmptyChannels(t *testing.T) {
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
	channels := []string{} // Empty channels
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with empty channels, but it passed")
	}

	// Check that the error message mentions channels
	foundChannelError := false
	for _, err := range result.Errors {
		if err == "channels list cannot be empty" {
			foundChannelError = true
			break
		}
	}
	if !foundChannelError {
		t.Errorf("Expected error about empty channels, but got: %v", result.Errors)
	}
}

// TestValidateSetup_InvalidMatrixJobCount verifies validation fails with invalid matrix job count
func TestValidateSetup_InvalidMatrixJobCount(t *testing.T) {
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
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 25 // Invalid: exceeds limit

	result := validator.ValidateSetup(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with invalid matrix job count, but it passed")
	}

	// Check that the error message mentions matrix_job_count
	foundMatrixError := false
	for _, err := range result.Errors {
		if err == "matrix_job_count cannot exceed 20 (GitHub Actions limit), got 25" {
			foundMatrixError = true
			break
		}
	}
	if !foundMatrixError {
		t.Errorf("Expected error about matrix_job_count exceeding limit, but got: %v", result.Errors)
	}
}

// TestValidateSetup_PartialNtfyConfiguration verifies validation fails with incomplete ntfy config
func TestValidateSetup_PartialNtfyConfiguration(t *testing.T) {
	// Set up all required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	
	// Set only NTFY_SERVER_URL without NTFY_TOPIC
	os.Setenv("NTFY_SERVER_URL", "https://ntfy.sh")
	os.Unsetenv("NTFY_TOPIC")
	
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
		os.Unsetenv("NTFY_SERVER_URL")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with partial ntfy config, but it passed")
	}

	// Check that the error message mentions missing NTFY_TOPIC
	foundNtfyError := false
	for _, err := range result.Errors {
		if err == "NTFY_SERVER_URL is set but NTFY_TOPIC is missing" {
			foundNtfyError = true
			break
		}
	}
	if !foundNtfyError {
		t.Errorf("Expected error about missing NTFY_TOPIC, but got: %v", result.Errors)
	}
}

// TestValidateSetup_PartialNtfyConfiguration_MissingServer verifies validation fails with only topic
func TestValidateSetup_PartialNtfyConfiguration_MissingServer(t *testing.T) {
	// Set up all required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	
	// Set only NTFY_TOPIC without NTFY_SERVER_URL
	os.Unsetenv("NTFY_SERVER_URL")
	os.Setenv("NTFY_TOPIC", "my-topic")
	
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
		os.Unsetenv("NTFY_TOPIC")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with partial ntfy config, but it passed")
	}

	// Check that the error message mentions missing NTFY_SERVER_URL
	foundNtfyError := false
	for _, err := range result.Errors {
		if err == "NTFY_TOPIC is set but NTFY_SERVER_URL is missing" {
			foundNtfyError = true
			break
		}
	}
	if !foundNtfyError {
		t.Errorf("Expected error about missing NTFY_SERVER_URL, but got: %v", result.Errors)
	}
}

// TestValidateSetup_CompleteNtfyConfiguration verifies validation passes with complete ntfy config
func TestValidateSetup_CompleteNtfyConfiguration(t *testing.T) {
	// Set up all required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	
	// Set complete ntfy configuration
	os.Setenv("NTFY_SERVER_URL", "https://ntfy.sh")
	os.Setenv("NTFY_TOPIC", "my-topic")
	os.Setenv("NTFY_TOKEN", "my-token")
	
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
		os.Unsetenv("NTFY_SERVER_URL")
		os.Unsetenv("NTFY_TOPIC")
		os.Unsetenv("NTFY_TOKEN")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	if !result.Valid {
		t.Errorf("Expected validation to pass with complete ntfy config, but got errors: %v", result.Errors)
	}
}

// TestValidateSetup_NoNotificationServices verifies validation passes without notification services
func TestValidateSetup_NoNotificationServices(t *testing.T) {
	// Set up all required environment variables
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	os.Setenv("GOFILE_API_KEY", "test-gofile-key")
	os.Setenv("FILESTER_API_KEY", "test-filester-key")
	
	// Clear all notification service environment variables
	os.Unsetenv("DISCORD_WEBHOOK_URL")
	os.Unsetenv("NTFY_SERVER_URL")
	os.Unsetenv("NTFY_TOPIC")
	os.Unsetenv("NTFY_TOKEN")
	
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GOFILE_API_KEY")
		os.Unsetenv("FILESTER_API_KEY")
	}()

	validator := NewConfigValidator()
	channels := []string{"chaturbate:user1"}
	matrixJobCount := 5

	result := validator.ValidateSetup(channels, matrixJobCount)

	// Validation should pass even without notification services (they're optional)
	if !result.Valid {
		t.Errorf("Expected validation to pass without notification services, but got errors: %v", result.Errors)
	}
}

// TestValidateSetup_MultipleErrors verifies all errors are collected in setup validation
func TestValidateSetup_MultipleErrors(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	validator := NewConfigValidator()
	channels := []string{} // Empty channels
	matrixJobCount := 0    // Invalid matrix job count

	result := validator.ValidateSetup(channels, matrixJobCount)

	if result.Valid {
		t.Error("Expected validation to fail with multiple errors, but it passed")
	}

	// Should have at least 6 errors: empty channels, matrix_job_count too low,
	// missing GITHUB_TOKEN, GITHUB_REPOSITORY, GOFILE_API_KEY, FILESTER_API_KEY
	if len(result.Errors) < 6 {
		t.Errorf("Expected at least 6 errors, but got %d: %v", len(result.Errors), result.Errors)
	}
}
