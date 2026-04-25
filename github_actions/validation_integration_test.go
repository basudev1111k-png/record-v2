package github_actions

import (
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

// TestValidationIntegration_ParseConfigFailure verifies that ParseGitHubActionsModeConfig
// returns an error when validation fails, which will cause the CLI app to exit with non-zero code.
//
// Requirements: 5.10
func TestValidationIntegration_ParseConfigFailure(t *testing.T) {
	// Clear environment variables to trigger validation failure
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")
	os.Unsetenv("MATRIX_JOB_ID")
	os.Unsetenv("CHANNELS")

	// Create a test CLI context with flags
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "mode", Value: "github-actions"},
			&cli.StringFlag{Name: "matrix-job-id", Value: ""},
			&cli.StringFlag{Name: "session-id", Value: ""},
			&cli.StringFlag{Name: "channels", Value: ""},
			&cli.BoolFlag{Name: "max-quality", Value: false},
			&cli.BoolFlag{Name: "validate-setup", Value: false},
		},
		Action: func(c *cli.Context) error {
			_, err := ParseGitHubActionsModeConfig(c)
			return err
		},
	}

	// Run the app with test arguments - this should fail validation
	err := app.Run([]string{"test", "--mode", "github-actions"})

	// Verify that an error is returned
	if err == nil {
		t.Error("Expected ParseGitHubActionsModeConfig to return an error with invalid config, but got nil")
	}

	// Verify the error message is descriptive
	if !strings.Contains(err.Error(), "required") && !strings.Contains(err.Error(), "validation") {
		t.Errorf("Expected error message to be descriptive, but got: %v", err)
	}
}

// TestValidationIntegration_ValidateSetupFailure verifies that ValidateSetupMode
// returns an error when validation fails.
//
// Requirements: 5.10
func TestValidationIntegration_ValidateSetupFailure(t *testing.T) {
	// Clear environment variables to trigger validation failure
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("GOFILE_API_KEY")
	os.Unsetenv("FILESTER_API_KEY")

	// Create a test CLI context
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "channels", Value: ""},
			&cli.BoolFlag{Name: "validate-setup", Value: true},
		},
		Action: func(c *cli.Context) error {
			return ValidateSetupMode(c)
		},
	}

	err := app.Run([]string{"test", "--validate-setup"})

	// Verify that an error is returned
	if err == nil {
		t.Error("Expected ValidateSetupMode to return an error with invalid config, but got nil")
	}

	// Verify the error message is descriptive
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected error message to mention 'validation failed', but got: %v", err)
	}

	// Verify the error message includes error count
	if !strings.Contains(err.Error(), "errors found") {
		t.Errorf("Expected error message to include error count, but got: %v", err)
	}
}

// TestValidationIntegration_SuccessfulValidation verifies that validation passes
// when all required configuration is present.
//
// Requirements: 5.10
func TestValidationIntegration_SuccessfulValidation(t *testing.T) {
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

	// Create a test CLI context
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "channels", Value: "channel1,channel2"},
			&cli.BoolFlag{Name: "validate-setup", Value: true},
		},
		Action: func(c *cli.Context) error {
			return ValidateSetupMode(c)
		},
	}

	err := app.Run([]string{"test", "--channels", "channel1,channel2", "--validate-setup"})

	// Verify that no error is returned
	if err != nil {
		t.Errorf("Expected ValidateSetupMode to succeed with valid config, but got error: %v", err)
	}
}

// TestValidationIntegration_ErrorMessageFormat verifies that error messages
// follow a consistent format and provide actionable information.
//
// Requirements: 5.10
func TestValidationIntegration_ErrorMessageFormat(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func()
		cleanupEnv     func()
		channels       string
		matrixJobCount string
		expectError    bool
		errorContains  []string
	}{
		{
			name: "Missing all secrets",
			setupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			cleanupEnv:     func() {},
			channels:       "channel1",
			matrixJobCount: "5",
			expectError:    true,
			errorContains: []string{
				"validation failed",
				"errors found",
			},
		},
		{
			name: "Empty channels",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "test-token")
				os.Setenv("GITHUB_REPOSITORY", "owner/repo")
				os.Setenv("GOFILE_API_KEY", "test-key")
				os.Setenv("FILESTER_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			channels:       "",
			matrixJobCount: "5",
			expectError:    true,
			errorContains: []string{
				"validation failed",
			},
		},
		{
			name: "Invalid matrix job count",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "test-token")
				os.Setenv("GITHUB_REPOSITORY", "owner/repo")
				os.Setenv("GOFILE_API_KEY", "test-key")
				os.Setenv("FILESTER_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GITHUB_REPOSITORY")
				os.Unsetenv("GOFILE_API_KEY")
				os.Unsetenv("FILESTER_API_KEY")
			},
			channels:       "channel1",
			matrixJobCount: "25",
			expectError:    true,
			errorContains: []string{
				"validation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Set MATRIX_JOB_COUNT environment variable for the test
			if tt.matrixJobCount != "" {
				os.Setenv("MATRIX_JOB_COUNT", tt.matrixJobCount)
				defer os.Unsetenv("MATRIX_JOB_COUNT")
			}

			// Create a test CLI context
			app := &cli.App{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "channels", Value: tt.channels},
					&cli.BoolFlag{Name: "validate-setup", Value: true},
				},
				Action: func(c *cli.Context) error {
					return ValidateSetupMode(c)
				},
			}

			args := []string{"test", "--validate-setup"}
			if tt.channels != "" {
				args = append(args, "--channels", tt.channels)
			}

			err := app.Run(args)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				} else {
					// Verify error message contains expected strings
					errMsg := err.Error()
					for _, expected := range tt.errorContains {
						if !strings.Contains(errMsg, expected) {
							t.Errorf("Expected error message to contain '%s', but got: %v", expected, errMsg)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
