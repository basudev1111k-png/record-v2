# Requirements Document

## Introduction

This document specifies requirements for implementing smart GitHub Actions workarounds to enable GoondVR (a livestream recording service) to run continuously despite GitHub Actions' 6-hour maximum job execution time and other platform limitations. The system will use an auto-restart chain pattern, state persistence strategies, external storage integration, and monitoring capabilities to achieve near-continuous operation within GitHub Actions' constraints.

## Glossary

- **Runner**: A GitHub Actions virtual machine that executes workflow jobs
- **Workflow**: A GitHub Actions automated process defined in YAML that runs jobs
- **Chain_Manager**: The component responsible for triggering the next workflow run before the current one times out
- **State_Persister**: The component that saves and restores application state between workflow runs
- **Storage_Uploader**: The component that uploads completed recordings to external storage services
- **Monitor**: The component that tracks system health and sends notifications
- **Session**: A single workflow run instance (maximum 5.5 hours of operation)
- **Transition**: The period when one workflow run ends and the next begins
- **Cache_Store**: GitHub Actions cache storage (10 GB limit per repository)
- **Artifact_Store**: GitHub Actions artifact storage for completed files
- **External_Storage**: Third-party storage services (S3-compatible, GitHub Releases, Google Drive, Dropbox)
- **GoondVR**: The livestream recording application being adapted for GitHub Actions
- **Matrix_Job**: A GitHub Actions job instance created from a matrix strategy, handling one channel independently
- **Matrix_Coordinator**: The component that orchestrates multiple matrix jobs and manages shared state
- **Gofile**: An external file hosting service used for storing completed recordings
- **Filester**: An external file hosting service used for storing completed recordings (10 GB per file, 45-day retention)
- **Database_Manager**: The component that maintains organized records of uploaded videos in the repository
- **Quality_Selector**: The component that determines and sets the optimal recording quality

## Requirements

### Requirement 1: Auto-Restart Chain Pattern

**User Story:** As a user, I want the system to automatically restart itself before the 6-hour timeout, so that recording continues without manual intervention.

#### Acceptance Criteria

1. WHEN a workflow run has been active for 5.5 hours, THE Chain_Manager SHALL trigger a new workflow run via the GitHub API
2. THE Chain_Manager SHALL pass the current session state to the new workflow run using workflow_dispatch inputs
3. WHEN the new workflow run starts, THE Chain_Manager SHALL verify the previous run is still active before taking over
4. THE Workflow SHALL terminate gracefully after 5.5 hours of operation
5. THE Chain_Manager SHALL include a unique session identifier in each workflow dispatch to prevent duplicate chains
6. IF the GitHub API call to trigger the next workflow fails, THEN THE Chain_Manager SHALL retry up to 3 times with exponential backoff
7. THE Chain_Manager SHALL log all chain transitions with timestamps and session identifiers

### Requirement 2: State Persistence Between Runs

**User Story:** As a user, I want the system to preserve configuration and recording state across workflow restarts, so that no data or settings are lost during transitions.

#### Acceptance Criteria

1. WHEN a workflow run begins, THE State_Persister SHALL restore configuration files from Cache_Store
2. WHEN a workflow run begins, THE State_Persister SHALL restore partial recordings from Cache_Store
3. WHEN a workflow run is about to end, THE State_Persister SHALL save all configuration files to Cache_Store
4. WHEN a workflow run is about to end, THE State_Persister SHALL save all partial recordings to Cache_Store
5. THE State_Persister SHALL use cache keys that include the session identifier to prevent conflicts
6. THE State_Persister SHALL verify cache integrity using checksums before restoring state
7. IF cache restoration fails, THEN THE State_Persister SHALL initialize with default configuration and log the failure
8. THE State_Persister SHALL maintain a state manifest file listing all cached files with their timestamps and sizes

### Requirement 3: Completed Recording Upload

**User Story:** As a user, I want completed recordings automatically uploaded to external storage, so that they are preserved and the runner disk space is freed.

#### Acceptance Criteria

1. WHEN a recording is completed, THE Storage_Uploader SHALL upload the file to the configured External_Storage
2. THE Storage_Uploader SHALL support Gofile and Filester as primary upload targets (see Requirement 14)
3. THE Storage_Uploader SHALL support GitHub Releases as an upload target
4. THE Storage_Uploader SHALL support S3-compatible storage (R2, Backblaze B2) as upload targets
5. THE Storage_Uploader SHALL support Google Drive as an upload target
6. THE Storage_Uploader SHALL support Dropbox as an upload target
7. WHEN an upload succeeds, THE Storage_Uploader SHALL delete the local file to free disk space
8. IF an upload fails, THEN THE Storage_Uploader SHALL retry up to 3 times with exponential backoff
9. IF all upload retries fail, THEN THE Storage_Uploader SHALL save the file to Artifact_Store as a fallback
10. THE Storage_Uploader SHALL log all upload operations with file names, sizes, destinations, and status
11. THE Storage_Uploader SHALL verify uploaded file integrity using checksums

### Requirement 4: Disk Space Management

**User Story:** As a user, I want the system to monitor and manage disk space usage, so that the runner does not exceed the 14 GB limit and fail.

#### Acceptance Criteria

1. THE Monitor SHALL check available disk space every 5 minutes
2. WHEN disk usage exceeds 10 GB, THE Monitor SHALL trigger immediate upload of completed recordings
3. WHEN disk usage exceeds 12 GB, THE Monitor SHALL pause new recording starts until space is freed
4. WHEN disk usage exceeds 13 GB, THE Monitor SHALL stop the oldest active recording and upload it immediately
5. THE Monitor SHALL log disk usage statistics with each check
6. THE Monitor SHALL send a notification when disk space management actions are taken

### Requirement 5: Configuration Management

**User Story:** As a user, I want to configure the system through environment variables and workflow inputs, so that I can customize behavior without modifying code.

#### Acceptance Criteria

1. THE Workflow SHALL accept a list of channels to monitor as a workflow input
2. THE Workflow SHALL accept external storage credentials as encrypted secrets
3. THE Workflow SHALL accept Gofile API key as an encrypted secret
4. THE Workflow SHALL accept Filester API key as an encrypted secret
5. THE Workflow SHALL accept polling interval configuration as a workflow input
6. THE Workflow SHALL accept recording quality settings as workflow inputs (see Requirement 16)
7. THE Workflow SHALL accept notification webhook URLs as encrypted secrets
8. THE Workflow SHALL accept matrix job count as a workflow input (maximum 20)
9. THE Workflow SHALL validate all configuration inputs before starting operation
10. IF configuration validation fails, THEN THE Workflow SHALL fail with a descriptive error message
11. THE Workflow SHALL validate that required API keys for Gofile and Filester are present when dual upload is enabled

### Requirement 6: Health Monitoring and Notifications

**User Story:** As a user, I want to receive notifications about system status and failures, so that I can respond to issues promptly.

#### Acceptance Criteria

1. WHEN a workflow run starts, THE Monitor SHALL send a notification with the session identifier
2. WHEN a Matrix_Job starts, THE Monitor SHALL send a notification with the matrix job identifier and assigned channel
3. WHEN a workflow run ends normally, THE Monitor SHALL send a notification with session statistics
4. WHEN a workflow run fails, THE Monitor SHALL send a notification with error details
5. WHEN a Matrix_Job fails, THE Monitor SHALL send a notification identifying the failed job and affected channel
6. WHEN a chain transition occurs, THE Monitor SHALL send a notification with transition status
7. WHEN a recording starts, THE Monitor SHALL send a notification with channel, timestamp, and quality settings
8. WHEN a recording completes, THE Monitor SHALL send a notification with file size, quality, and upload status to both Gofile and Filester
9. THE Monitor SHALL support Discord webhooks as a notification target
10. THE Monitor SHALL support ntfy as a notification target
11. THE Monitor SHALL include a health check endpoint that returns current system status
12. THE Monitor SHALL detect gaps in recording coverage during transitions and report them
13. THE Monitor SHALL aggregate status from all active Matrix_Jobs for overall system health reporting

### Requirement 7: Graceful Shutdown

**User Story:** As a user, I want the system to shut down gracefully before the timeout, so that recordings are properly closed and state is saved.

#### Acceptance Criteria

1. WHEN the workflow run reaches 5.4 hours, THE Workflow SHALL initiate graceful shutdown
2. WHEN graceful shutdown begins, THE Workflow SHALL stop accepting new recording starts
3. WHEN graceful shutdown begins, THE Workflow SHALL allow active recordings to continue for up to 5 minutes
4. WHEN graceful shutdown begins, THE Workflow SHALL trigger the next workflow run via Chain_Manager
5. WHEN all recordings are closed or the 5-minute grace period expires, THE Workflow SHALL save state via State_Persister
6. WHEN state is saved, THE Workflow SHALL upload any completed recordings via Storage_Uploader
7. THE Workflow SHALL complete within 5.5 hours total runtime

### Requirement 8: Recovery from Failures

**User Story:** As a user, I want the system to recover automatically from transient failures, so that operation continues despite temporary issues.

#### Acceptance Criteria

1. WHEN a GitHub API call fails with a transient error, THE Workflow SHALL retry the operation
2. WHEN cache restoration fails, THE Workflow SHALL initialize with default state and continue
3. WHEN an upload fails, THE Storage_Uploader SHALL use the fallback Artifact_Store
4. WHEN a workflow run fails to start, THE Chain_Manager SHALL detect the gap and trigger a new run
5. THE Workflow SHALL log all recovery actions with timestamps and error details
6. WHEN a recording stream fails, THE Workflow SHALL retry the recording after the configured interval

### Requirement 9: Optimization for GitHub Actions Environment

**User Story:** As a user, I want the system optimized for GitHub Actions constraints, so that it operates efficiently within platform limits.

#### Acceptance Criteria

1. THE Workflow SHALL reduce polling interval to 5 minutes during periods with no active recordings
2. THE Workflow SHALL prioritize channels based on configured priority levels
3. WHERE a recording size limit is configured, THE Workflow SHALL split recordings at the specified size
4. WHERE a recording quality limit is configured, THE Workflow SHALL use lower quality settings to reduce disk usage
5. THE Workflow SHALL use GitHub Actions cache compression to maximize the 10 GB cache limit
6. THE Workflow SHALL clean up temporary files immediately after they are no longer needed
7. THE Workflow SHALL use incremental cache updates to minimize cache save time

### Requirement 10: Deployment and Setup

**User Story:** As a user, I want simple deployment instructions, so that I can set up the system quickly without deep GitHub Actions knowledge.

#### Acceptance Criteria

1. THE documentation SHALL provide a step-by-step setup guide
2. THE documentation SHALL include a template workflow YAML file
3. THE documentation SHALL list all required secrets and their purposes
4. THE documentation SHALL explain how to configure each external storage option
5. THE documentation SHALL include troubleshooting steps for common issues
6. THE Workflow SHALL include a setup validation mode that checks configuration without starting recordings
7. THE documentation SHALL clearly state the limitations and expected gaps during transitions

### Requirement 11: Monitoring Dashboard

**User Story:** As a user, I want to view current system status through a simple interface, so that I can monitor operation without checking logs.

#### Acceptance Criteria

1. THE Workflow SHALL generate a status badge showing current operational state
2. THE Workflow SHALL update a status file in the repository with current session information
3. THE status file SHALL include the current session identifier, start time, and active recordings count
4. THE status file SHALL include the number of active Matrix_Jobs and their assigned channels
5. THE status file SHALL include disk usage statistics
6. THE status file SHALL include the timestamp of the last successful chain transition
7. THE status file SHALL include the count of successful uploads to Gofile and Filester
8. THE status file SHALL be updated every 5 minutes during operation
9. THE Workflow SHALL commit the status file to the repository on each update
10. THE status file SHALL include per-Matrix_Job status showing channel, recording state, and last activity timestamp

### Requirement 12: Cost Management

**User Story:** As a user, I want to stay within GitHub's free tier limits, so that I can run the system without incurring charges.

#### Acceptance Criteria

1. THE documentation SHALL calculate expected GitHub Actions minutes usage per month
2. THE documentation SHALL calculate expected cache storage usage
3. THE documentation SHALL calculate expected artifact storage usage
4. THE documentation SHALL provide recommendations for staying within free tier limits
5. THE Workflow SHALL include optional cost-saving modes that reduce resource usage
6. WHERE cost-saving mode is enabled, THE Workflow SHALL reduce polling frequency to 10 minutes
7. WHERE cost-saving mode is enabled, THE Workflow SHALL limit concurrent recordings to 2 channels

### Requirement 13: Matrix Job Parallel Channel Recording

**User Story:** As a user, I want to record up to 20 channels simultaneously using GitHub Actions matrix strategy, so that multiple streams can be captured in parallel without interference.

#### Acceptance Criteria

1. THE Workflow SHALL use GitHub Actions matrix strategy to create up to 20 parallel Matrix_Jobs
2. THE Workflow SHALL assign exactly one channel to each Matrix_Job
3. THE Matrix_Jobs SHALL execute concurrently on separate Runners
4. WHEN a Matrix_Job starts, THE Matrix_Job SHALL operate independently with its own 5.5-hour lifecycle
5. WHEN a Matrix_Job reaches 5.5 hours, THE Matrix_Job SHALL trigger its own replacement via the auto-restart chain pattern
6. THE Matrix_Coordinator SHALL distribute channel assignments across Matrix_Jobs using workflow inputs
7. THE Matrix_Coordinator SHALL use GitHub Actions cache with job-specific cache keys to prevent conflicts between Matrix_Jobs
8. IF a Matrix_Job fails, THEN the other Matrix_Jobs SHALL continue operation without interruption
9. THE Matrix_Coordinator SHALL maintain a shared configuration cache accessible to all Matrix_Jobs
10. THE Workflow SHALL validate that the number of channels does not exceed 20 before creating Matrix_Jobs

### Requirement 14: Dual Upload to Gofile and Filester

**User Story:** As a user, I want completed recordings uploaded to both Gofile and Filester, so that I have redundant storage and multiple access options.

#### Acceptance Criteria

1. WHEN a recording is completed, THE Storage_Uploader SHALL upload the file to both Gofile and Filester
2. THE Storage_Uploader SHALL retrieve the Gofile server address from https://api.gofile.io/servers before uploading
3. WHEN uploading to Gofile, THE Storage_Uploader SHALL send a POST request to https://{server}.gofile.io/uploadFile with multipart/form-data encoding
4. WHEN uploading to Filester, THE Storage_Uploader SHALL send a POST request to https://u1.filester.me/api/v1/upload with multipart/form-data encoding
5. THE Storage_Uploader SHALL include the Gofile API key from GitHub secrets in the Gofile upload request as a Bearer token
6. THE Storage_Uploader SHALL include the Filester API key from GitHub secrets in the Filester upload request as a Bearer token in the Authorization header
7. THE Storage_Uploader SHALL verify both uploads succeed by checking HTTP response status codes (200 for Gofile, 200 for Filester)
8. WHEN both uploads succeed, THE Storage_Uploader SHALL extract the download URLs from the JSON responses (Gofile: "url" field, Filester: "url" field)
9. WHEN both uploads succeed, THE Storage_Uploader SHALL delete the local file to free disk space
10. IF either upload fails, THEN THE Storage_Uploader SHALL retry that specific upload up to 3 times with exponential backoff
11. IF either upload fails after 3 retries, THEN THE Storage_Uploader SHALL fall back to Artifact_Store and log the failure
12. THE Storage_Uploader SHALL execute Gofile and Filester uploads in parallel to minimize upload time
13. THE Storage_Uploader SHALL log both upload URLs with timestamps and file metadata
14. WHERE a recording file exceeds 10 GB, THE Storage_Uploader SHALL upload the full file to Gofile and split the file into 10 GB chunks for Filester
15. WHEN uploading split files to Filester, THE Storage_Uploader SHALL create a folder for the recording and upload all chunks to that folder
16. THE Storage_Uploader SHALL store all Filester chunk URLs in the database for split recordings

### Requirement 15: Database Organization in Repository

**User Story:** As a user, I want uploaded video links organized in a structured database within the repository, so that I can easily find and access recordings by channel and date.

#### Acceptance Criteria

1. THE Database_Manager SHALL create a database directory in the repository root
2. THE Database_Manager SHALL organize video metadata using the path structure database/{site}/{channel_username}/{YYYY-MM-DD}.json
3. WHEN a recording is successfully uploaded, THE Database_Manager SHALL create or update the JSON file for that channel and date
4. THE Database_Manager SHALL store recording metadata as a JSON array containing objects with timestamp, duration_seconds, file_size_bytes, quality, gofile_url, filester_url, filester_chunks (for split files), session_id, and matrix_job fields
5. THE Database_Manager SHALL use ISO 8601 format for timestamp fields
6. THE Database_Manager SHALL include the workflow run identifier in the session_id field
7. THE Database_Manager SHALL include the matrix job identifier in the matrix_job field
8. WHEN updating a database file, THE Database_Manager SHALL perform a git pull operation before modifying the file
9. WHEN updating a database file, THE Database_Manager SHALL append the new recording entry to the existing array
10. WHEN a database file is updated, THE Database_Manager SHALL commit the changes with a descriptive message including channel name and timestamp
11. WHEN a database file is committed, THE Database_Manager SHALL push the commit to the remote repository
12. IF a git push fails due to conflicts, THEN THE Database_Manager SHALL retry with git pull, merge, and push up to 3 times
13. THE Database_Manager SHALL use atomic git operations to prevent database corruption from concurrent Matrix_Job updates
14. THE Database_Manager SHALL validate JSON structure before committing to prevent malformed database entries

### Requirement 16: Maximum Quality Recording

**User Story:** As a user, I want recordings captured at the highest available quality up to 4K 60fps, so that video quality is maximized within bandwidth and storage constraints.

#### Acceptance Criteria

1. THE Quality_Selector SHALL attempt to record at 2160p (4K) resolution as the first priority
2. THE Quality_Selector SHALL attempt to record at 60 frames per second as the first priority
3. WHEN 4K 60fps is not available, THE Quality_Selector SHALL fall back to 1080p 60fps
4. WHEN 1080p 60fps is not available, THE Quality_Selector SHALL fall back to 720p 60fps
5. WHEN 720p 60fps is not available, THE Quality_Selector SHALL select the highest available quality
6. THE Quality_Selector SHALL set the resolution flag to 2160 when attempting 4K recording
7. THE Quality_Selector SHALL set the framerate flag to 60 when attempting 60fps recording
8. WHEN a recording starts, THE Quality_Selector SHALL log the actual quality being recorded
9. THE Quality_Selector SHALL include the actual recorded quality in the format "{resolution}p{framerate}" in database entries
10. THE Quality_Selector SHALL override any existing resolution or framerate configuration settings with maximum quality settings
11. THE Quality_Selector SHALL detect the available quality options from the stream before starting recording

### Requirement 17: Matrix Job State Coordination

**User Story:** As a user, I want matrix jobs to coordinate state and database updates safely, so that concurrent operations do not cause data loss or corruption.

#### Acceptance Criteria

1. THE Matrix_Coordinator SHALL assign unique cache keys to each Matrix_Job using the pattern "state-{session_id}-{matrix_job_id}"
2. THE Matrix_Coordinator SHALL provide a shared configuration cache key accessible to all Matrix_Jobs
3. WHEN a Matrix_Job writes to Cache_Store, THE Matrix_Job SHALL use only its assigned cache key
4. WHEN a Matrix_Job reads shared configuration, THE Matrix_Job SHALL use the shared configuration cache key
5. THE Matrix_Coordinator SHALL use git pull-commit-push sequences for all database updates to handle concurrent modifications
6. WHEN multiple Matrix_Jobs update the database simultaneously, THE Database_Manager SHALL use git's conflict resolution to merge changes
7. IF a Matrix_Job detects a git conflict, THEN THE Database_Manager SHALL retry the update operation with fresh data from git pull
8. THE Matrix_Coordinator SHALL maintain a job registry in the shared cache listing all active Matrix_Jobs with their channel assignments
9. WHEN a Matrix_Job starts, THE Matrix_Job SHALL register itself in the job registry
10. WHEN a Matrix_Job completes, THE Matrix_Job SHALL remove itself from the job registry
11. THE Matrix_Coordinator SHALL use the job registry to detect and recover from failed Matrix_Jobs
12. THE Matrix_Coordinator SHALL ensure database updates are atomic at the file level to prevent partial writes
