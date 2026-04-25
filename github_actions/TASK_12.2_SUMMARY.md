# Task 12.2 Summary: Add Setup Validation Mode

## Overview

This task implements a `--validate-setup` flag that allows users to validate their GitHub Actions configuration and secrets without actually starting recordings. This is useful for testing the environment before running the full workflow.

## Implementation Details

### 1. Added `--validate-setup` Flag

**File:** `github_actions/github_actions_mode.go`

Added a new boolean flag to the GitHub Actions mode flags:

```go
&cli.BoolFlag{
    Name:  "validate-setup",
    Usage: "Validate configuration and secrets without starting recordings",
    Value: false,
}
```

### 2. Implemented `ValidateSetup()` Method

**File:** `github_actions/config_validator.go`

Created a comprehensive validation method that checks:

- **Required Environment Variables:**
  - `GITHUB_TOKEN` - GitHub API authentication token
  - `GITHUB_REPOSITORY` - GitHub repository identifier
  - `GOFILE_API_KEY` - Gofile upload API key
  - `FILESTER_API_KEY` - Filester upload API key

- **Workflow Inputs:**
  - Channels list is not empty
  - Matrix job count is between 1 and 20

- **Optional Notification Configuration:**
  - Discord webhook URL (optional)
  - Ntfy server and topic (both required if either is set)
  - Ntfy token (optional)

The method returns a `ValidationResult` with all validation errors found.

### 3. Implemented `ValidateSetupMode()` Function

**File:** `github_actions/github_actions_mode.go`

Created a function that:

1. Parses channels and matrix job count from flags/environment
2. Calls `ValidateSetup()` to perform comprehensive validation
3. Displays validation results with detailed error messages
4. Shows configuration summary with masked secrets
5. Returns error if validation fails, nil if all checks pass

### 4. Added `maskSecret()` Helper Function

**File:** `github_actions/github_actions_mode.go`

Created a helper function to safely display secrets in logs:

- Shows `<not set>` for empty secrets
- Shows `****` for secrets 8 characters or less
- Shows first 4 and last 4 characters with `****` in between for longer secrets

Example: `abcdefghijklmnop` → `abcd****mnop`

### 5. Integrated Validation into `ParseGitHubActionsModeConfig()`

**File:** `github_actions/github_actions_mode.go`

Modified the config parsing function to check for `--validate-setup` flag:

```go
// Check if we're in validate-setup mode
if c.Bool("validate-setup") {
    return nil, ValidateSetupMode(c)
}
```

When the flag is set, the function runs validation and exits without creating a `GitHubActionsMode` instance or starting recordings.

## Usage

### Command Line

```bash
# Validate setup without starting recordings
go run main.go \
  --mode github-actions \
  --validate-setup \
  --channels "channel1,channel2,channel3" \
  --matrix-job-id matrix-job-1

# Or with environment variables
export GITHUB_TOKEN="ghp_xxxx"
export GITHUB_REPOSITORY="owner/repo"
export GOFILE_API_KEY="xxxx"
export FILESTER_API_KEY="xxxx"
export CHANNELS="channel1,channel2"
export MATRIX_JOB_COUNT="5"

go run main.go --mode github-actions --validate-setup
```

### Example Output (Success)

```
=== GitHub Actions Setup Validation ===

Validation Results:
-------------------
✅ All validation checks passed!

Configuration Summary:
  - Channels: [channel1 channel2 channel3]
  - Matrix Job Count: 5
  - GITHUB_TOKEN: ghp_****xxxx
  - GITHUB_REPOSITORY: owner/repo
  - GOFILE_API_KEY: test****key1
  - FILESTER_API_KEY: test****key2

Optional Notification Configuration:
  - Discord Webhook: https****hook
  - Ntfy Server: https://ntfy.sh
  - Ntfy Topic: my-topic
  - Ntfy Token: tk_****ken1

✅ Setup validation completed successfully!
You can now run the workflow without --validate-setup to start recordings.
```

### Example Output (Failure)

```
=== GitHub Actions Setup Validation ===

Validation Results:
-------------------
❌ Validation failed with 4 error(s):

  1. channels list cannot be empty
  2. GITHUB_TOKEN environment variable is required (GitHub API authentication token)
  3. GOFILE_API_KEY environment variable is required (Gofile upload API key)
  4. FILESTER_API_KEY environment variable is required (Filester upload API key)

Please fix the above errors before running the workflow.
```

## Testing

### Unit Tests Added

**File:** `github_actions/config_validator_test.go`

Added comprehensive tests for the `ValidateSetup()` method:

1. `TestValidateSetup_AllValid` - Validates with all required configuration
2. `TestValidateSetup_MissingAllSecrets` - Validates with all secrets missing
3. `TestValidateSetup_EmptyChannels` - Validates with empty channels list
4. `TestValidateSetup_InvalidMatrixJobCount` - Validates with invalid matrix job count
5. `TestValidateSetup_PartialNtfyConfiguration` - Validates with incomplete ntfy config
6. `TestValidateSetup_PartialNtfyConfiguration_MissingServer` - Validates with only ntfy topic
7. `TestValidateSetup_CompleteNtfyConfiguration` - Validates with complete ntfy config
8. `TestValidateSetup_NoNotificationServices` - Validates without notification services
9. `TestValidateSetup_MultipleErrors` - Validates that all errors are collected

**File:** `github_actions/github_actions_mode_test.go`

Added test for the `maskSecret()` helper function:

1. `TestMaskSecret` - Tests secret masking with various input lengths

### Test Results

All tests pass successfully:

```
=== RUN   TestValidateSetup_AllValid
--- PASS: TestValidateSetup_AllValid (0.00s)
=== RUN   TestValidateSetup_MissingAllSecrets
--- PASS: TestValidateSetup_MissingAllSecrets (0.00s)
=== RUN   TestValidateSetup_EmptyChannels
--- PASS: TestValidateSetup_EmptyChannels (0.00s)
=== RUN   TestValidateSetup_InvalidMatrixJobCount
--- PASS: TestValidateSetup_InvalidMatrixJobCount (0.00s)
=== RUN   TestValidateSetup_PartialNtfyConfiguration
--- PASS: TestValidateSetup_PartialNtfyConfiguration (0.00s)
=== RUN   TestValidateSetup_PartialNtfyConfiguration_MissingServer
--- PASS: TestValidateSetup_PartialNtfyConfiguration_MissingServer (0.00s)
=== RUN   TestValidateSetup_CompleteNtfyConfiguration
--- PASS: TestValidateSetup_CompleteNtfyConfiguration (0.00s)
=== RUN   TestValidateSetup_NoNotificationServices
--- PASS: TestValidateSetup_NoNotificationServices (0.00s)
=== RUN   TestValidateSetup_MultipleErrors
--- PASS: TestValidateSetup_MultipleErrors (0.00s)
PASS
```

## Requirements Satisfied

✅ **Requirement 10.6:** The Workflow SHALL include a setup validation mode that checks configuration without starting recordings

### Validation Checks Implemented

1. ✅ All required secrets are present (GITHUB_TOKEN, GITHUB_REPOSITORY, GOFILE_API_KEY, FILESTER_API_KEY)
2. ✅ Channels list is not empty
3. ✅ Matrix job count is within valid range (1-20)
4. ✅ Optional notification configuration is validated (Discord, ntfy)
5. ✅ Exits without starting recordings
6. ✅ Provides clear error messages for validation failures
7. ✅ Shows configuration summary with masked secrets

## Files Modified

1. `github_actions/config_validator.go` - Added `ValidateSetup()` method
2. `github_actions/github_actions_mode.go` - Added `--validate-setup` flag, `ValidateSetupMode()` function, and `maskSecret()` helper
3. `github_actions/config_validator_test.go` - Added 9 new test cases for setup validation
4. `github_actions/github_actions_mode_test.go` - Added test for `maskSecret()` function

## Integration with Workflow

The `--validate-setup` flag can be used in GitHub Actions workflows to validate configuration before starting the actual recording workflow:

```yaml
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Validate Setup
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GOFILE_API_KEY: ${{ secrets.GOFILE_API_KEY }}
          FILESTER_API_KEY: ${{ secrets.FILESTER_API_KEY }}
        run: |
          go run main.go \
            --mode github-actions \
            --validate-setup \
            --channels "${{ inputs.channels }}"
```

## Benefits

1. **Early Error Detection:** Catches configuration issues before starting expensive workflow runs
2. **Clear Error Messages:** Provides specific guidance on what needs to be fixed
3. **Security:** Masks secrets in output to prevent accidental exposure
4. **Comprehensive:** Validates all required and optional configuration
5. **Fast Feedback:** Runs quickly without starting actual recordings

## Next Steps

This implementation completes task 12.2. The next task (12.3) will handle validation failures by failing the workflow with descriptive error messages and logging validation errors.
