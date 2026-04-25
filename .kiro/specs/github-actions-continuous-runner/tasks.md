# Implementation Plan: GitHub Actions Continuous Runner

## Overview

This implementation adapts GoondVR to run continuously on GitHub Actions using an auto-restart chain pattern, matrix-based parallel recording, dual external storage uploads (Gofile and Filester), and state persistence. The system will handle up to 20 concurrent channels with automatic workflow transitions every 5.5 hours.

## Tasks

- [x] 1. Create GitHub Actions workflow infrastructure
  - Create `.github/workflows/continuous-runner.yml` with workflow_dispatch trigger
  - Define workflow inputs: session_state, channels, matrix_job_count
  - Configure matrix strategy with dynamic job count (max 20)
  - Set timeout to 330 minutes (5.5 hours)
  - Add cache restore/save steps for state persistence
  - Configure fail-fast: false for matrix jobs
  - _Requirements: 5.1, 5.2, 5.8, 13.1, 13.2, 13.3_

- [x] 2. Implement Chain Manager component
  - [x] 2.1 Create chain_manager.go with ChainManager struct
    - Implement GenerateSessionID() to create unique session identifiers
    - Implement TriggerNextRun() to call GitHub API workflow_dispatch endpoint
    - Build GitHub API payload with session state and workflow inputs
    - Use GITHUB_TOKEN from environment for authentication
    - _Requirements: 1.1, 1.2, 1.5_
  
  - [x] 2.2 Add runtime monitoring and chain triggering logic
    - Implement MonitorRuntime() to check elapsed time every minute
    - Trigger next workflow run at 5.5 hours (19,800 seconds)
    - Pass current session state via workflow inputs
    - Verify previous run is still active before new run takes over
    - _Requirements: 1.1, 1.3, 1.4_
  
  - [x] 2.3 Add retry logic with exponential backoff
    - Implement RetryWithBackoff() helper function
    - Retry GitHub API calls up to 3 times on transient errors
    - Use exponential backoff: 1s, 2s, 4s delays
    - Log all retry attempts with error details
    - _Requirements: 1.6, 1.7_

- [ ]* 2.4 Write unit tests for Chain Manager
    - Test session ID uniqueness
    - Test GitHub API payload construction
    - Test retry logic with mock API responses
    - Test timing calculations for 5.5-hour trigger
    - _Requirements: 1.1, 1.5, 1.6_

- [x] 3. Implement State Persister component
  - [x] 3.1 Create state_persister.go with StatePersister struct
    - Define StateManifest and FileEntry structs
    - Implement SaveState() to persist config and recordings to cache
    - Implement RestoreState() to retrieve state from cache
    - Use cache keys: `state-{session_id}-{matrix_job_id}`
    - Use shared config key: `shared-config-latest`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  
  - [x] 3.2 Add cache integrity verification
    - Implement VerifyIntegrity() to check checksums
    - Calculate SHA-256 checksums for all cached files
    - Validate checksums before restoring state
    - Log cache misses and integrity failures
    - _Requirements: 2.6, 2.8_
  
  - [x] 3.3 Add state manifest management
    - Create manifest file listing all cached files
    - Include timestamps, sizes, and checksums in manifest
    - Update manifest on each state save operation
    - _Requirements: 2.8_
  
  - [x] 3.4 Handle cache restoration failures
    - Initialize with default configuration on cache miss
    - Log warnings for missing cache entries
    - Continue operation with fresh state
    - _Requirements: 2.7_

- [ ]* 3.5 Write unit tests for State Persister
    - Test cache key generation for different session IDs
    - Test checksum calculation and verification
    - Test state manifest serialization/deserialization
    - Test handling of missing cache entries
    - _Requirements: 2.5, 2.6, 2.7, 2.8_

- [x] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Implement Storage Uploader component
  - [x] 5.1 Create storage_uploader.go with StorageUploader struct
    - Define UploadResult struct with GofileURL, FilesterURL, FilesterChunks fields
    - Implement GetGofileServer() to retrieve optimal server from https://api.gofile.io/servers
    - Parse JSON response and extract server address
    - _Requirements: 3.1, 3.2, 14.2_
  
  - [x] 5.2 Implement Gofile upload functionality
    - Implement UploadToGofile() with multipart/form-data encoding
    - Send POST request to https://{server}.gofile.io/uploadFile
    - Include Gofile API key as Bearer token in Authorization header
    - Parse JSON response and extract download URL from "url" field
    - Verify HTTP 200 response status
    - _Requirements: 14.1, 14.3, 14.5, 14.7, 14.8_
  
  - [x] 5.3 Implement Filester upload functionality
    - Implement UploadToFilester() with multipart/form-data encoding
    - Send POST request to https://u1.filester.me/api/v1/upload
    - Include Filester API key as Bearer token in Authorization header
    - Parse JSON response and extract download URL from "url" field
    - Verify HTTP 200 response status
    - _Requirements: 14.1, 14.4, 14.6, 14.7, 14.8_
  
  - [x] 5.4 Add file splitting for Filester 10 GB limit
    - Check file size before upload
    - Split files > 10 GB into 10 GB chunks
    - Create folder on Filester for split recordings
    - Upload all chunks to the folder
    - Return array of chunk URLs
    - _Requirements: 14.14, 14.15, 14.16_
  
  - [x] 5.5 Implement parallel dual upload
    - Implement UploadRecording() to coordinate both uploads
    - Execute Gofile and Filester uploads in parallel using goroutines
    - Wait for both uploads to complete
    - Return combined UploadResult with both URLs
    - _Requirements: 14.1, 14.12_
  
  - [x] 5.6 Add retry logic and fallback handling
    - Retry individual uploads up to 3 times with exponential backoff
    - Track which upload failed (Gofile or Filester)
    - Implement FallbackToArtifacts() for GitHub Artifacts upload
    - Fall back to artifacts if both uploads fail after retries
    - Log all upload operations with status
    - _Requirements: 3.8, 14.10, 14.11_
  
  - [x] 5.7 Add local file cleanup
    - Delete local file after successful dual upload
    - Verify both Gofile and Filester uploads succeeded before deletion
    - Log file deletion operations
    - _Requirements: 3.7, 14.9_
  
  - [x] 5.8 Add upload integrity verification
    - Calculate file checksum before upload
    - Verify uploaded file integrity using checksums
    - Log verification results
    - _Requirements: 3.11_

- [ ]* 5.9 Write unit tests for Storage Uploader
    - Test Gofile server selection logic
    - Test multipart form data construction
    - Test parallel upload coordination
    - Test file splitting for Filester 10 GB limit
    - Test fallback to artifacts on failure
    - _Requirements: 14.1, 14.2, 14.10, 14.11, 14.14_

- [ ] 6. Implement Database Manager component
  - [x] 6.1 Create database_manager.go with DatabaseManager struct
    - Define RecordingMetadata struct with all required fields
    - Implement GetDatabasePath() to generate path: `database/{site}/{channel}/{YYYY-MM-DD}.json`
    - Create directory structure if it doesn't exist
    - _Requirements: 15.1, 15.2_
  
  - [x] 6.2 Implement atomic database updates
    - Implement AtomicUpdate() with git pull-commit-push sequence
    - Use mutex to prevent concurrent git operations
    - Perform git pull before modifying file
    - Append new recording entry to JSON array
    - Commit with descriptive message including channel and timestamp
    - Push to remote repository
    - _Requirements: 15.8, 15.9, 15.10, 15.13_
  
  - [x] 6.3 Add git conflict resolution
    - Detect git push failures due to conflicts
    - Retry with git pull, merge, and push up to 3 times
    - Log all conflict resolution attempts
    - _Requirements: 15.12_
  
  - [x] 6.4 Implement AddRecording() method
    - Create or update JSON file for channel and date
    - Build RecordingMetadata with timestamp (ISO 8601), duration, file size, quality
    - Include gofile_url, filester_url, filester_chunks (if split)
    - Include session_id and matrix_job identifiers
    - Validate JSON structure before committing
    - _Requirements: 15.3, 15.4, 15.5, 15.6, 15.7, 15.14_

- [ ]* 6.5 Write unit tests for Database Manager
    - Test database path generation for different sites/channels
    - Test JSON structure validation
    - Test metadata serialization
    - Test concurrent update handling (mock git operations)
    - _Requirements: 15.2, 15.4, 15.13, 15.14_

- [x] 7. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 8. Implement Matrix Coordinator component
  - [x] 8.1 Create matrix_coordinator.go with MatrixCoordinator struct
    - Define MatrixJobInfo struct with JobID, Channel, StartTime, Status fields
    - Initialize job registry map with mutex for thread safety
    - _Requirements: 13.1, 13.2_
  
  - [x] 8.2 Implement channel assignment logic
    - Implement AssignChannels() to distribute channels across matrix jobs
    - Validate channel count does not exceed 20
    - Create JobAssignment for each matrix job
    - Ensure exactly one channel per matrix job
    - _Requirements: 13.2, 13.6, 13.10_
  
  - [x] 8.3 Implement job registry operations
    - Implement RegisterJob() to add matrix job to registry
    - Implement UnregisterJob() to remove matrix job from registry
    - Implement GetActiveJobs() to return all active jobs
    - Store registry in shared cache for cross-job visibility
    - _Requirements: 17.8, 17.9, 17.10_
  
  - [x] 8.4 Add failed job detection
    - Implement DetectFailedJobs() to identify stale jobs
    - Check for jobs that haven't reported in expected time
    - Return list of failed job IDs
    - _Requirements: 13.8, 17.11_
  
  - [x] 8.5 Implement cache key management
    - Assign unique cache keys per matrix job: `state-{session_id}-{matrix_job_id}`
    - Provide shared configuration cache key: `shared-config-latest`
    - Ensure matrix jobs use only their assigned cache keys
    - _Requirements: 13.7, 17.1, 17.2, 17.3, 17.4_

- [ ]* 8.6 Write unit tests for Matrix Coordinator
    - Test channel distribution across jobs
    - Test job registry operations
    - Test failed job detection logic
    - Test cache key generation
    - _Requirements: 13.2, 13.6, 13.10, 17.1, 17.8_

- [ ] 9. Implement Quality Selector component
  - [x] 9.1 Create quality_selector.go with QualitySelector struct
    - Define QualitySettings struct with Resolution, Framerate, Actual fields
    - Set preferred resolution to 2160 (4K)
    - Set preferred framerate to 60
    - _Requirements: 16.1, 16.2, 16.6, 16.7_
  
  - [x] 9.2 Implement quality selection logic
    - Implement SelectQuality() with fallback chain
    - Priority 1: 2160p @ 60fps
    - Priority 2: 1080p @ 60fps
    - Priority 3: 720p @ 60fps
    - Priority 4: Highest available
    - Return QualitySettings with selected resolution and framerate
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_
  
  - [x] 9.3 Add stream quality detection
    - Implement DetectAvailableQualities() to query stream
    - Parse available quality options from stream metadata
    - Return list of available qualities
    - _Requirements: 16.11_
  
  - [x] 9.4 Implement configuration override
    - Implement ApplyQualitySettings() to configure recording engine
    - Override existing resolution and framerate settings
    - Apply maximum quality settings to entity.ChannelConfig
    - Log actual quality being recorded
    - Format quality string as "{resolution}p{framerate}"
    - _Requirements: 16.8, 16.9, 16.10_

- [ ]* 9.5 Write unit tests for Quality Selector
    - Test quality fallback logic (4K60 → 1080p60 → 720p60)
    - Test quality string formatting
    - Test configuration override logic
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.9_

- [ ] 10. Implement Health Monitor component
  - [x] 10.1 Create health_monitor.go with HealthMonitor struct
    - Define SystemStatus struct with all required fields
    - Define MatrixJobStatus struct for per-job status
    - Initialize notifiers array (Discord, ntfy)
    - Set disk check interval to 5 minutes
    - _Requirements: 6.1, 6.2, 11.3, 11.4_
  
  - [x] 10.2 Implement disk space monitoring
    - Implement MonitorDiskSpace() to check usage every 5 minutes
    - Trigger immediate upload at 10 GB usage
    - Pause new recordings at 12 GB usage
    - Stop oldest recording at 13 GB usage
    - Log disk usage statistics with each check
    - Send notifications for disk management actions
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_
  
  - [x] 10.3 Implement notification system
    - Implement SendNotification() to send alerts via configured notifiers
    - Support Discord webhooks
    - Support ntfy notifications
    - Send notifications for workflow start/end
    - Send notifications for matrix job start/fail
    - Send notifications for chain transitions
    - Send notifications for recording start/complete with quality and upload status
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10_
  
  - [x] 10.4 Implement status file management
    - Implement UpdateStatusFile() to write current system status
    - Include session_id, start_time, active_recordings count
    - Include active_matrix_jobs array with per-job status
    - Include disk_usage_bytes and disk_total_bytes
    - Include last_chain_transition timestamp
    - Include gofile_uploads and filester_uploads counts
    - Update status file every 5 minutes
    - Commit status file to repository on each update
    - _Requirements: 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8, 11.9, 11.10_
  
  - [x] 10.5 Add recording gap detection
    - Implement DetectRecordingGaps() to identify gaps during transitions
    - Compare timestamps between workflow runs
    - Report gaps in notifications
    - _Requirements: 6.12_
  
  - [x] 10.6 Add matrix job health aggregation
    - Aggregate status from all active matrix jobs
    - Report overall system health
    - Include per-job status in system status
    - _Requirements: 6.13, 11.4, 11.10_

- [x] 11. Integrate components into main application
  - [x] 11.1 Create github_actions_mode.go for GitHub Actions integration
    - Add command-line flag `--mode github-actions`
    - Add flags for `--matrix-job-id`, `--session-id`, `--channels`, `--max-quality`
    - Parse workflow inputs from environment variables
    - Initialize all components (ChainManager, StatePersister, etc.)
    - _Requirements: 5.1, 5.2, 5.5, 5.6, 5.8_
  
  - [x] 11.2 Implement workflow lifecycle management
    - Restore state from cache on startup
    - Start Chain Manager runtime monitoring in background goroutine
    - Start Health Monitor disk space monitoring in background goroutine
    - Register matrix job with Matrix Coordinator
    - _Requirements: 2.1, 4.1, 13.4, 17.9_
  
  - [x] 11.3 Implement graceful shutdown logic
    - Detect 5.4-hour runtime threshold
    - Stop accepting new recording starts
    - Allow active recordings to continue for up to 5 minutes
    - Trigger next workflow run via Chain Manager
    - Save state via State Persister
    - Upload completed recordings via Storage Uploader
    - Unregister matrix job from Matrix Coordinator
    - Complete within 5.5 hours total runtime
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 17.10_
  
  - [x] 11.4 Wire recording completion to uploads and database
    - Hook into existing recording completion event
    - Call Storage Uploader to upload to Gofile and Filester
    - Call Database Manager to add recording metadata
    - Send notification via Health Monitor
    - Delete local file after successful upload
    - _Requirements: 3.1, 3.7, 6.7, 14.1, 15.3_
  
  - [x] 11.5 Apply maximum quality settings
    - Call Quality Selector to determine best quality
    - Apply quality settings to channel configuration
    - Override any existing quality settings
    - Log actual quality being recorded
    - _Requirements: 16.1, 16.2, 16.8, 16.10_

- [x] 12. Implement configuration validation
  - [x] 12.1 Add workflow input validation
    - Validate channels list is not empty
    - Validate matrix_job_count is between 1 and 20
    - Validate Gofile API key is present
    - Validate Filester API key is present
    - Validate polling interval is positive
    - _Requirements: 5.9, 5.11_
  
  - [x] 12.2 Add setup validation mode
    - Add `--validate-setup` flag
    - Check all required secrets are present
    - Check configuration is valid
    - Exit without starting recordings
    - _Requirements: 10.6_
  
  - [x] 12.3 Handle validation failures
    - Fail workflow with descriptive error message
    - Log validation errors
    - _Requirements: 5.10_

- [x] 13. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Add error recovery mechanisms
  - [x] 14.1 Implement GitHub API failure recovery
    - Retry API calls with exponential backoff
    - Log all retry attempts
    - Continue current workflow until timeout if chain trigger fails
    - _Requirements: 8.1, 8.5_
  
  - [x] 14.2 Implement cache restoration failure recovery
    - Initialize with default state on cache miss
    - Log warnings for missing cache
    - Continue operation with fresh state
    - _Requirements: 8.2_
  
  - [x] 14.3 Implement upload failure recovery
    - Use fallback Artifact_Store on upload failure
    - Log failure with file details
    - Send notification about fallback usage
    - Continue operation
    - _Requirements: 8.3_
  
  - [x] 14.4 Implement workflow start failure detection
    - Detect gaps in chain transitions
    - Log missing workflow runs
    - Send notification about chain gap
    - _Requirements: 8.4_
  
  - [x] 14.5 Implement recording stream failure recovery
    - Retry recording after configured interval
    - Log stream failures
    - Continue monitoring other channels
    - _Requirements: 8.6_

- [ ] 15. Add optimization features
  - [x] 15.1 Implement adaptive polling
    - Reduce polling interval to 5 minutes when no active recordings
    - Use normal interval when recordings are active
    - _Requirements: 9.1_
  
  - [x] 15.2 Implement cache compression
    - Use GitHub Actions cache compression
    - Maximize 10 GB cache limit
    - _Requirements: 9.5_
  
  - [x] 15.3 Implement incremental cache updates
    - Update cache incrementally instead of full saves
    - Minimize cache save time
    - _Requirements: 9.7_
  
  - [x] 15.4 Implement temporary file cleanup
    - Clean up temporary files immediately after use
    - Free disk space proactively
    - _Requirements: 9.6_
  
  - [x] 15.5 Add cost-saving mode
    - Add `--cost-saving` flag
    - Reduce polling frequency to 10 minutes
    - Limit concurrent recordings to 2 channels
    - _Requirements: 12.5, 12.6, 12.7_

- [ ] 16. Create documentation
  - [x] 16.1 Write setup guide
    - Document step-by-step setup instructions
    - List all required secrets and their purposes
    - Explain how to configure Gofile and Filester
    - Include troubleshooting steps for common issues
    - _Requirements: 10.1, 10.3, 10.4, 10.5_
  
  - [x] 16.2 Create template workflow YAML
    - Provide complete workflow file template
    - Include comments explaining each section
    - Document all workflow inputs
    - _Requirements: 10.2_
  
  - [x] 16.3 Document limitations and gaps
    - Clearly state 30-60 second recording gaps during transitions
    - Explain 6-hour job limit constraint
    - Document expected GitHub Actions minutes usage
    - Provide cost estimation for free tier
    - _Requirements: 10.7, 12.1, 12.2, 12.3, 12.4_
  
  - [ ] 16.4 Add operational guide
    - Document monitoring procedures
    - Explain status file format
    - Describe failure recovery procedures
    - Include cost management recommendations
    - _Requirements: 11.1, 12.4_

- [ ] 17. Final integration and testing
  - [ ] 17.1 Test end-to-end workflow
    - Deploy workflow to test repository
    - Configure with single test channel
    - Verify workflow runs for 5 minutes
    - Verify state is saved to cache
    - Verify next workflow is triggered
    - Verify transition completes successfully
    - _Requirements: 1.1, 1.2, 2.1, 2.3, 7.4_
  
  - [ ] 17.2 Test matrix job independence
    - Deploy workflow with 3 matrix jobs
    - Verify jobs run concurrently
    - Verify each job handles one channel
    - Verify job failure doesn't affect others
    - _Requirements: 13.3, 13.4, 13.8_
  
  - [ ] 17.3 Test dual upload functionality
    - Record test stream
    - Verify upload to both Gofile and Filester
    - Verify both URLs are returned
    - Verify local file is deleted
    - Verify database entry includes both URLs
    - _Requirements: 14.1, 14.8, 14.9, 15.3_
  
  - [ ] 17.4 Test database concurrent updates
    - Trigger multiple matrix jobs to complete recordings simultaneously
    - Verify all database updates are preserved
    - Verify no data loss from git conflicts
    - Verify JSON structure remains valid
    - _Requirements: 15.12, 15.13, 15.14_
  
  - [ ] 17.5 Test quality selection
    - Verify 4K 60fps is attempted first
    - Verify fallback to lower qualities works
    - Verify actual quality is logged and stored in database
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.9_
  
  - [ ] 17.6 Test graceful shutdown
    - Run workflow for 5.4 hours
    - Verify graceful shutdown initiates
    - Verify active recordings complete
    - Verify state is saved
    - Verify next workflow is triggered
    - Verify completion within 5.5 hours
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

- [ ] 18. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- The implementation uses Go as specified in the design document
- Matrix jobs operate independently with their own 5.5-hour lifecycle
- Dual uploads to Gofile and Filester provide storage redundancy
- Database updates use git operations for atomic concurrent modifications
- Quality selector maximizes recording quality up to 4K 60fps
- This solution exceeds GitHub's free tier limits for continuous 24/7 operation
