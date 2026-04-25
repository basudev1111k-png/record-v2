# Task 11.2 Summary: Implement Workflow Lifecycle Management

## Overview

This task implements the workflow lifecycle management for GitHub Actions mode, which orchestrates the startup sequence for a matrix job in the continuous runner system.

## Implementation Details

### StartWorkflowLifecycle Method

Added the `StartWorkflowLifecycle()` method to `GitHubActionsMode` that performs the following operations in sequence:

1. **Restore State from Cache**
   - Calls `StatePersister.RestoreState()` to retrieve cached configuration and recordings
   - Handles cache misses gracefully (expected for first run)
   - Logs warnings for cache restoration failures but continues with default state
   - Implements Requirement 2.1 (restore configuration from cache)

2. **Register Matrix Job**
   - Determines the assigned channel for this matrix job using `GetAssignedChannel()`
   - Registers the job with the Matrix Coordinator using `RegisterJob()`
   - Implements Requirement 17.9 (register matrix job with coordinator)
   - Implements Requirement 13.4 (matrix job operates independently)

3. **Start Chain Manager Monitoring**
   - Launches Chain Manager runtime monitoring in a background goroutine
   - Monitors elapsed time and triggers next workflow run at 5.5 hours
   - Provides a state provider function that returns current session state
   - Implements Requirement 2.1 (chain manager runtime monitoring)

4. **Start Health Monitor Disk Space Monitoring**
   - Launches Health Monitor disk space monitoring in a background goroutine
   - Checks disk usage every 5 minutes
   - Takes progressive actions based on usage thresholds (10 GB, 12 GB, 13 GB)
   - Implements Requirement 4.1 (disk space monitoring)

5. **Send Workflow Start Notification**
   - Sends a notification when the workflow starts
   - Includes matrix job ID, assigned channel, and session ID
   - Continues even if notification fails

### Bug Fix: ChainManager SessionID

Fixed an issue where the ChainManager's internal sessionID was not being set when a sessionID was provided to `NewGitHubActionsMode()`. The fix ensures that when a sessionID is provided (not auto-generated), it is explicitly set on the ChainManager instance.

```go
if gam.SessionID == "" {
    gam.SessionID = gam.ChainManager.GenerateSessionID()
} else {
    // Set the ChainManager's sessionID to match the provided sessionID
    gam.ChainManager.sessionID = gam.SessionID
}
```

## Test Coverage

Created comprehensive test suite in `github_actions_mode_test.go`:

### Test Cases

1. **TestStartWorkflowLifecycle**
   - Tests the complete workflow lifecycle startup
   - Verifies state restoration (cache miss scenario)
   - Verifies matrix job registration
   - Verifies background goroutines start correctly
   - Verifies session ID is set correctly

2. **TestStartWorkflowLifecycle_CacheMiss**
   - Tests graceful handling of cache miss (no cached state)
   - Verifies system continues with default state
   - Verifies matrix job is still registered despite cache miss

3. **TestStartWorkflowLifecycle_InvalidMatrixJobID**
   - Tests error handling for invalid matrix job ID format
   - Verifies appropriate error is returned
   - Ensures system fails fast on invalid configuration

4. **TestStartWorkflowLifecycle_BackgroundMonitoring**
   - Tests that background goroutines start and run
   - Verifies Chain Manager has not triggered next run immediately
   - Verifies context cancellation stops background goroutines

5. **TestGetAssignedChannel**
   - Tests channel assignment logic for matrix jobs
   - Verifies first job gets first channel, second gets second, etc.
   - Tests error handling for invalid job ID format
   - Tests error handling for job index out of range

6. **TestStartWorkflowLifecycle_WithCachedState**
   - Tests lifecycle with existing cached state
   - Verifies state is restored successfully
   - Verifies matrix job registration after state restoration

### Test Results

All tests pass successfully:
```
PASS: TestStartWorkflowLifecycle (0.21s)
PASS: TestStartWorkflowLifecycle_CacheMiss (0.00s)
PASS: TestStartWorkflowLifecycle_InvalidMatrixJobID (0.00s)
PASS: TestStartWorkflowLifecycle_BackgroundMonitoring (0.21s)
PASS: TestStartWorkflowLifecycle_WithCachedState (0.08s)
PASS: TestGetAssignedChannel (0.00s)
```

## Requirements Satisfied

- **Requirement 2.1**: Restore configuration files from Cache_Store on workflow start
- **Requirement 4.1**: Monitor disk space every 5 minutes
- **Requirement 13.4**: Matrix jobs operate independently with their own lifecycle
- **Requirement 17.9**: Register matrix job with Matrix Coordinator on startup

## Files Modified

1. **github_actions/github_actions_mode.go**
   - Added `StartWorkflowLifecycle()` method
   - Fixed ChainManager sessionID initialization bug

2. **github_actions/github_actions_mode_test.go** (new file)
   - Added comprehensive test suite for workflow lifecycle management
   - Added tests for channel assignment logic
   - Added tests for error handling scenarios

## Integration Points

The `StartWorkflowLifecycle()` method integrates with:

1. **StatePersister**: Restores cached state on startup
2. **MatrixCoordinator**: Registers the matrix job for coordination
3. **ChainManager**: Starts runtime monitoring for auto-restart chain
4. **HealthMonitor**: Starts disk space monitoring and sends notifications
5. **StorageUploader**: Passed to HealthMonitor for emergency upload actions

## Usage Example

```go
// Create GitHubActionsMode instance
gam, err := NewGitHubActionsMode(matrixJobID, sessionID, channels, maxQuality)
if err != nil {
    log.Fatalf("Failed to create GitHub Actions mode: %v", err)
}
defer gam.Cancel()

// Start workflow lifecycle
err = gam.StartWorkflowLifecycle(configDir, recordingsDir)
if err != nil {
    log.Fatalf("Failed to start workflow lifecycle: %v", err)
}

// Workflow is now running with:
// - State restored from cache
// - Matrix job registered
// - Chain Manager monitoring runtime
// - Health Monitor checking disk space
```

## Next Steps

The next task (11.3) will implement graceful shutdown logic to:
- Detect 5.4-hour runtime threshold
- Stop accepting new recording starts
- Allow active recordings to complete
- Save state and upload completed recordings
- Unregister matrix job from coordinator

## Notes

- Background goroutines are managed via the context (`gam.ctx`)
- Calling `gam.Cancel()` stops all background monitoring
- Cache restoration failures are handled gracefully (system continues with default state)
- Notification failures do not stop workflow startup
- The system is designed to be resilient to transient failures
