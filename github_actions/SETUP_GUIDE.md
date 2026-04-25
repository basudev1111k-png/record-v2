# GitHub Actions Continuous Runner - Setup Guide

This guide provides step-by-step instructions for setting up the GoondVR GitHub Actions Continuous Runner. This system enables continuous livestream recording despite GitHub Actions' 6-hour job timeout by using an auto-restart chain pattern.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Required Secrets Configuration](#required-secrets-configuration)
4. [Workflow Configuration](#workflow-configuration)
5. [Storage Provider Setup](#storage-provider-setup)
6. [Notification Setup](#notification-setup)
7. [Running the Workflow](#running-the-workflow)
8. [Monitoring and Status](#monitoring-and-status)
9. [Cost Considerations](#cost-considerations)
10. [Troubleshooting](#troubleshooting)
11. [Advanced Configuration](#advanced-configuration)

---

## Prerequisites

Before setting up the continuous runner, ensure you have:

- **GitHub Account**: With access to GitHub Actions
- **Repository Access**: Fork or clone the GoondVR repository
- **Storage Accounts**: At least one of the following:
  - Gofile API account (recommended)
  - Filester API account (recommended)
  - GitHub repository for releases
  - S3-compatible storage (R2, Backblaze B2)
  - Google Drive account
  - Dropbox account
- **Notification Services** (optional but recommended):
  - Discord webhook URL
  - ntfy server access

---

## Quick Start

For users who want to get started immediately:

1. **Fork the repository** to your GitHub account
2. **Configure secrets** (see [Required Secrets](#required-secrets-configuration))
3. **Trigger the workflow** manually with your channel list
4. **Monitor status** via the status file in your repository

The workflow is already configured in `.github/workflows/continuous-runner.yml` and ready to use.

---

## Required Secrets Configuration

GitHub Actions secrets store sensitive information like API keys. Configure these in your repository settings.

### Accessing Repository Secrets

1. Navigate to your repository on GitHub
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret** for each secret below

### Required Secrets

#### 1. GOFILE_API_KEY (Recommended)

**Purpose**: Enables automatic upload of completed recordings to Gofile storage.

**How to obtain**:
1. Visit [Gofile.io](https://gofile.io)
2. Create a free account or log in
3. Navigate to your account settings
4. Generate an API key
5. Copy the API key

**Configuration**:
- **Name**: `GOFILE_API_KEY`
- **Value**: Your Gofile API key (e.g., `abc123def456...`)

**Notes**:
- Gofile offers unlimited storage on free tier
- Files are retained indefinitely on free accounts
- Upload speed may vary based on server load

#### 2. FILESTER_API_KEY (Recommended)

**Purpose**: Provides redundant storage for recordings alongside Gofile.

**How to obtain**:
1. Visit [Filester.me](https://filester.me)
2. Create an account or log in
3. Navigate to API settings
4. Generate an API key
5. Copy the API key

**Configuration**:
- **Name**: `FILESTER_API_KEY`
- **Value**: Your Filester API key (e.g., `xyz789abc123...`)

**Notes**:
- Maximum file size: 10 GB per file
- Files larger than 10 GB are automatically split into chunks
- Retention period: 45 days
- The system uploads to both Gofile and Filester for redundancy

#### 3. GITHUB_TOKEN (Automatic)

**Purpose**: Enables workflow dispatch, cache access, and repository operations.

**Configuration**: 
- **No action required** - GitHub automatically provides this token
- The token is available as `${{ secrets.GITHUB_TOKEN }}` in workflows
- Permissions are automatically configured for the workflow

**Notes**:
- This token is scoped to the repository
- It expires when the workflow run completes
- Used for triggering the auto-restart chain

#### 4. DISCORD_WEBHOOK_URL (Optional)

**Purpose**: Sends notifications about workflow events to Discord.

**How to obtain**:
1. Open Discord and navigate to your server
2. Go to **Server Settings** → **Integrations** → **Webhooks**
3. Click **New Webhook**
4. Name it (e.g., "GoondVR Notifications")
5. Select the channel for notifications
6. Click **Copy Webhook URL**

**Configuration**:
- **Name**: `DISCORD_WEBHOOK_URL`
- **Value**: Your Discord webhook URL (e.g., `https://discord.com/api/webhooks/...`)

**Notification Events**:
- Workflow start/end
- Matrix job start/failure
- Recording start/completion
- Chain transitions
- Disk space warnings
- Upload status

#### 5. NTFY_TOKEN (Optional)

**Purpose**: Sends push notifications via ntfy service.

**How to obtain**:
1. Visit [ntfy.sh](https://ntfy.sh) or your self-hosted ntfy server
2. Create an account (if using authentication)
3. Generate an access token
4. Copy the token

**Configuration**:
- **Name**: `NTFY_TOKEN`
- **Value**: Your ntfy access token

**Notes**:
- Can be used alongside or instead of Discord
- Supports mobile push notifications
- Can use public ntfy.sh or self-hosted instance

### Verifying Secrets

After adding secrets, verify they appear in your repository settings:
- Navigate to **Settings** → **Secrets and variables** → **Actions**
- You should see all configured secrets listed (values are hidden)
- Secrets are available to all workflows in the repository

---

## Workflow Configuration

The workflow is pre-configured in `.github/workflows/continuous-runner.yml`. You can customize behavior through workflow inputs when triggering manually.

### Workflow Inputs

When triggering the workflow manually, you'll be prompted for these inputs:

#### 1. channels (Required)

**Format**: Comma-separated list of channels in `site:username` format

**Examples**:
- Single channel: `chaturbate:username1`
- Multiple channels: `chaturbate:username1,stripchat:username2,chaturbate:username3`

**Supported Sites**:
- `chaturbate` - Chaturbate.com
- `stripchat` - Stripchat.com

**Notes**:
- Maximum 20 channels (GitHub Actions concurrent job limit)
- Each channel runs in an independent matrix job
- Channels are case-sensitive

#### 2. matrix_job_count (Optional)

**Purpose**: Controls the number of parallel matrix jobs

**Default**: `5`

**Range**: 1-20

**Recommendations**:
- Use `5` for 5 or fewer channels
- Use `10` for 10 or fewer channels
- Use `20` for maximum parallelism (up to 20 channels)
- In cost-saving mode, automatically limited to `2`

**Notes**:
- Each matrix job handles one channel independently
- More jobs = more GitHub Actions minutes consumed
- Free tier allows 20 concurrent jobs

#### 3. session_state (Optional)

**Purpose**: Passes state from previous workflow run for continuity

**Default**: Empty (first run)

**Format**: JSON string containing session state

**Notes**:
- **Do not set manually** - automatically managed by the auto-restart chain
- Used internally for workflow transitions
- Contains session ID, partial recordings, and configuration

#### 4. cost_saving (Optional)

**Purpose**: Enables cost-saving mode to reduce GitHub Actions minutes usage

**Default**: `false`

**Options**: `true` or `false`

**When enabled**:
- Polling interval increased to 10 minutes (from 5 minutes)
- Maximum concurrent channels limited to 2
- Matrix job count automatically capped at 2

**Recommendations**:
- Enable for free tier usage
- Disable for maximum recording coverage
- Consider for channels with infrequent streams

---

## Storage Provider Setup

The system supports multiple storage providers. Configure at least one for completed recordings.

### Gofile (Recommended)

**Advantages**:
- Unlimited storage on free tier
- No file size limits
- Indefinite retention
- Fast upload speeds

**Setup**:
1. Create account at [Gofile.io](https://gofile.io)
2. Generate API key in account settings
3. Add `GOFILE_API_KEY` secret to repository
4. No additional configuration needed

**API Details**:
- Server selection: Automatic (retrieves optimal server)
- Upload endpoint: `https://{server}.gofile.io/uploadFile`
- Authentication: Bearer token in Authorization header

### Filester (Recommended)

**Advantages**:
- Redundant storage alongside Gofile
- Good upload speeds
- Simple API

**Limitations**:
- 10 GB per file limit (automatically handled via splitting)
- 45-day retention period

**Setup**:
1. Create account at [Filester.me](https://filester.me)
2. Generate API key in account settings
3. Add `FILESTER_API_KEY` secret to repository
4. System automatically splits files > 10 GB

**API Details**:
- Upload endpoint: `https://u1.filester.me/api/v1/upload`
- Authentication: Bearer token in Authorization header
- Chunking: Automatic for files > 10 GB

### GitHub Releases (Fallback)

**Purpose**: Fallback storage when primary uploads fail

**Setup**:
- No additional configuration needed
- Uses `GITHUB_TOKEN` automatically
- Files saved as release artifacts

**Limitations**:
- 2 GB per file limit
- Consumes repository storage quota
- Manual cleanup required

### S3-Compatible Storage (Advanced)

**Supported Providers**:
- Cloudflare R2
- Backblaze B2
- AWS S3
- MinIO

**Setup** (requires code modification):
1. Add S3 credentials as secrets:
   - `S3_ACCESS_KEY_ID`
   - `S3_SECRET_ACCESS_KEY`
   - `S3_BUCKET_NAME`
   - `S3_ENDPOINT_URL`
2. Modify `storage_uploader.go` to enable S3 uploads
3. Configure bucket permissions for uploads

### Google Drive (Advanced)

**Setup** (requires code modification):
1. Create Google Cloud project
2. Enable Google Drive API
3. Create OAuth 2.0 credentials
4. Add credentials as secrets
5. Modify `storage_uploader.go` to enable Drive uploads

### Dropbox (Advanced)

**Setup** (requires code modification):
1. Create Dropbox app
2. Generate access token
3. Add `DROPBOX_ACCESS_TOKEN` secret
4. Modify `storage_uploader.go` to enable Dropbox uploads

---

## Notification Setup

Notifications keep you informed about workflow status and recording events.

### Discord Notifications

**Setup**:
1. Create Discord webhook (see [Required Secrets](#4-discord_webhook_url-optional))
2. Add `DISCORD_WEBHOOK_URL` secret
3. Notifications automatically sent to configured channel

**Notification Types**:
- ✅ Workflow started
- ✅ Workflow completed
- ⚠️ Workflow failed
- 🎬 Recording started
- ✅ Recording completed
- 📤 Upload status (Gofile + Filester)
- 🔄 Chain transition
- 💾 Disk space warnings

**Cooldown**: 5 minutes per notification type (prevents spam)

### ntfy Notifications

**Setup**:
1. Choose ntfy server:
   - Public: [ntfy.sh](https://ntfy.sh)
   - Self-hosted: Your own ntfy instance
2. Create topic (e.g., `goondvr-notifications`)
3. Generate access token (if using authentication)
4. Add `NTFY_TOKEN` secret
5. Configure topic in workflow environment variables

**Mobile Apps**:
- iOS: [ntfy app](https://apps.apple.com/app/ntfy/id1625396347)
- Android: [ntfy app](https://play.google.com/store/apps/details?id=io.heckel.ntfy)

**Advantages**:
- Push notifications to mobile devices
- No Discord server required
- Self-hosting option for privacy

### Disabling Notifications

To disable notifications:
- Simply don't configure `DISCORD_WEBHOOK_URL` or `NTFY_TOKEN` secrets
- The system will continue to operate without sending notifications
- Status information still available in status file and workflow logs

---

## Running the Workflow

### First Run

1. **Navigate to Actions tab** in your GitHub repository
2. **Select "GoondVR Continuous Runner"** workflow
3. **Click "Run workflow"** button
4. **Configure inputs**:
   - **channels**: Enter your channel list (e.g., `chaturbate:username1,stripchat:username2`)
   - **matrix_job_count**: Set to number of channels (max 20)
   - **cost_saving**: Enable if using free tier
   - **session_state**: Leave empty for first run
5. **Click "Run workflow"** to start

### Monitoring Progress

**Workflow Logs**:
- Click on the running workflow in the Actions tab
- View logs for each matrix job
- Check for errors or warnings

**Status File**:
- Located at `status.json` in repository root
- Updated every 5 minutes
- Contains current system state

**Notifications**:
- Receive real-time updates via Discord/ntfy
- Monitor recording starts and completions
- Get alerts for failures or issues

### Auto-Restart Chain

The workflow automatically restarts itself every 5.5 hours:

1. **At 5.4 hours**: Graceful shutdown begins
2. **At 5.5 hours**: New workflow triggered via GitHub API
3. **Transition**: 30-60 second gap in recording
4. **New workflow**: Restores state and continues

**What happens during transition**:
- Active recordings are closed gracefully
- State saved to GitHub Actions cache
- Completed recordings uploaded to storage
- New workflow dispatched with session state
- New workflow restores state and resumes

**Expected gaps**: 30-60 seconds per transition (every 5.5 hours)

### Stopping the Workflow

To stop the continuous runner:

1. **Cancel current workflow run**:
   - Go to Actions tab
   - Click on running workflow
   - Click "Cancel workflow"
2. **Prevent auto-restart**:
   - The chain will not trigger if workflow is cancelled
   - No manual cleanup needed
3. **Verify all jobs stopped**:
   - Check that all matrix jobs are cancelled
   - Verify no new workflow runs are triggered

---

## Monitoring and Status

### Status File

**Location**: `status.json` in repository root

**Update Frequency**: Every 5 minutes

**Contents**:
```json
{
  "session_id": "run-20240115-143000-abc",
  "start_time": "2024-01-15T14:30:00Z",
  "active_recordings": 3,
  "active_matrix_jobs": [
    {
      "job_id": "1",
      "channel": "chaturbate:username1",
      "recording_state": "recording",
      "last_activity": "2024-01-15T15:00:00Z"
    }
  ],
  "disk_usage_bytes": 5368709120,
  "disk_total_bytes": 15032385536,
  "last_chain_transition": "2024-01-15T14:30:00Z",
  "gofile_uploads": 10,
  "filester_uploads": 10
}
```

**Interpreting Status**:
- `active_recordings`: Number of currently recording streams
- `active_matrix_jobs`: Per-job status with channel assignments
- `disk_usage_bytes`: Current disk usage (monitor for capacity issues)
- `gofile_uploads` / `filester_uploads`: Count of successful uploads

### Database Organization

**Location**: `database/` directory in repository

**Structure**:
```
database/
├── chaturbate/
│   ├── username1/
│   │   ├── 2024-01-15.json
│   │   └── 2024-01-16.json
│   └── username2/
│       └── 2024-01-15.json
└── stripchat/
    └── username3/
        └── 2024-01-15.json
```

**JSON Format**:
```json
[
  {
    "timestamp": "2024-01-15T14:30:00Z",
    "duration_seconds": 3600,
    "file_size_bytes": 2147483648,
    "quality": "2160p60",
    "gofile_url": "https://gofile.io/d/abc123",
    "filester_url": "https://filester.me/file/xyz789",
    "filester_chunks": [],
    "session_id": "run-20240115-143000-abc",
    "matrix_job": "1"
  }
]
```

**Accessing Recordings**:
1. Navigate to `database/{site}/{channel}/{date}.json`
2. Find the recording by timestamp
3. Use `gofile_url` or `filester_url` to download
4. For split files, use `filester_chunks` array

### Workflow Logs

**Accessing Logs**:
1. Go to Actions tab
2. Click on workflow run
3. Select matrix job to view logs
4. Expand steps to see detailed output

**Key Log Sections**:
- **Validate inputs**: Configuration validation
- **Restore state**: Cache restoration status
- **Start recording**: Recording initialization
- **Upload**: Upload progress and results
- **Save state**: Cache save operations

### Health Monitoring

**Disk Space Monitoring**:
- Checked every 5 minutes
- Thresholds:
  - 10 GB: Trigger immediate uploads
  - 12 GB: Pause new recordings
  - 13 GB: Emergency stop oldest recording

**Notification Events**:
- Workflow lifecycle (start/end/fail)
- Recording events (start/complete)
- Upload status (success/failure)
- Disk space warnings
- Chain transitions

---

## Cost Considerations

### GitHub Actions Free Tier

**Limits**:
- 2,000 minutes per month
- 20 concurrent jobs
- 10 GB cache storage
- 500 MB artifact storage

### Usage Calculation

**24/7 Operation**:
- 720 hours/month = 43,200 minutes
- With 5 matrix jobs: 216,000 minutes/month
- **Exceeds free tier by 107x**

**Cost-Saving Mode**:
- 2 concurrent channels
- 10-minute polling
- Estimated: ~86,400 minutes/month
- **Still exceeds free tier by 43x**

### Recommendations

**For Free Tier Users**:

1. **Limited Hours Operation**:
   - Run 8 hours/day: ~14,400 minutes/month (within free tier)
   - Schedule during peak streaming hours
   - Use workflow scheduling or manual triggers

2. **Single Channel**:
   - 1 channel 24/7: ~43,200 minutes/month
   - Requires paid plan or reduced hours

3. **Paid Plan**:
   - GitHub Team: $4/user/month + $0.008/minute
   - GitHub Enterprise: Custom pricing

**Cost Optimization Strategies**:

1. **Enable cost-saving mode**: Reduces polling frequency
2. **Limit concurrent channels**: Use fewer matrix jobs
3. **Schedule recordings**: Only run during specific hours
4. **Use self-hosted runners**: No minute limits (requires own infrastructure)

### Alternative Platforms

If GitHub Actions costs are prohibitive, consider:

- **AWS Free Tier**: 750 hours/month EC2 (first 12 months)
- **Google Cloud Free Tier**: $300 credit (first 90 days)
- **Azure Free Tier**: $200 credit (first 30 days)
- **Oracle Cloud Free Tier**: Always-free compute instances
- **Self-hosted**: Run on your own hardware

---

## Troubleshooting

### Common Issues

#### 1. Workflow Fails to Start

**Symptoms**:
- Workflow doesn't appear in Actions tab
- "Run workflow" button doesn't work

**Solutions**:
- Verify workflow file exists at `.github/workflows/continuous-runner.yml`
- Check workflow YAML syntax (use YAML validator)
- Ensure you have write access to the repository
- Check GitHub Actions is enabled in repository settings

#### 2. Cache Restoration Fails

**Symptoms**:
- Warning: "No cached state found"
- Workflow starts with default configuration

**Solutions**:
- **First run**: This is expected - no cache exists yet
- **Subsequent runs**: Check cache storage quota (10 GB limit)
- Verify cache keys are correct in workflow YAML
- Check workflow logs for cache restoration errors

**Note**: Cache misses are handled gracefully - workflow continues with defaults

#### 3. Upload Failures

**Symptoms**:
- Recordings not appearing in Gofile/Filester
- Upload errors in workflow logs
- Files saved to GitHub Artifacts instead

**Solutions**:

**Gofile Issues**:
- Verify `GOFILE_API_KEY` secret is set correctly
- Check API key is valid (test at gofile.io)
- Review Gofile server selection in logs
- Check network connectivity in workflow logs

**Filester Issues**:
- Verify `FILESTER_API_KEY` secret is set correctly
- Check API key is valid (test at filester.me)
- For files > 10 GB, verify chunking is working
- Check upload endpoint is accessible

**General Upload Issues**:
- Check file size (ensure within limits)
- Verify network connectivity
- Review retry attempts in logs
- Check fallback to GitHub Artifacts worked

#### 4. Chain Doesn't Restart

**Symptoms**:
- Workflow stops after 5.5 hours
- No new workflow run triggered
- Recording stops permanently

**Solutions**:
- Verify `GITHUB_TOKEN` has correct permissions
- Check GitHub API rate limits (view in logs)
- Review Chain Manager logs for errors
- Manually trigger new workflow with previous session state

**Manual Recovery**:
1. Note the session ID from stopped workflow
2. Trigger new workflow manually
3. Set `session_state` input to previous session ID
4. Workflow will restore state and continue

#### 5. Disk Space Exhausted

**Symptoms**:
- Workflow fails with "No space left on device"
- Recordings stop unexpectedly
- Upload operations fail

**Solutions**:
- Check disk usage in status file
- Verify uploads are completing successfully
- Review disk space monitoring logs
- Reduce number of concurrent recordings
- Enable more aggressive upload triggers

**Prevention**:
- Monitor disk usage in status file
- Ensure uploads complete before disk fills
- Use cost-saving mode for fewer concurrent recordings

#### 6. Matrix Job Failures

**Symptoms**:
- One or more matrix jobs fail
- Some channels stop recording
- Other jobs continue normally

**Solutions**:
- Check failed job logs for specific errors
- Verify channel name is correct
- Check if channel is actually online
- Review stream availability
- Manually restart failed job if needed

**Note**: Matrix jobs are independent - one failure doesn't affect others

#### 7. Database Update Conflicts

**Symptoms**:
- Git push failures in logs
- Database entries missing
- Merge conflicts in database files

**Solutions**:
- Check git operation logs
- Verify repository permissions
- Review concurrent update handling
- Database Manager automatically retries with merge

**Note**: System handles conflicts automatically via git pull-merge-push

#### 8. Quality Selection Issues

**Symptoms**:
- Recordings not at expected quality
- Quality lower than 4K 60fps
- Quality inconsistent across recordings

**Solutions**:
- Check stream actually offers 4K 60fps
- Review Quality Selector logs
- Verify quality detection is working
- Check bandwidth limitations
- Review actual recorded quality in database entries

**Note**: System automatically falls back to highest available quality

### Getting Help

If you encounter issues not covered here:

1. **Check workflow logs**: Most issues have detailed error messages
2. **Review status file**: Shows current system state
3. **Check notifications**: May contain error details
4. **Search GitHub Issues**: Others may have encountered similar issues
5. **Open GitHub Issue**: Provide workflow logs and configuration details

### Debug Mode

To enable verbose logging:

1. Add repository variable `ACTIONS_STEP_DEBUG` = `true`
2. Add repository variable `ACTIONS_RUNNER_DEBUG` = `true`
3. Re-run workflow
4. Review detailed debug logs

---

## Advanced Configuration

### Custom Polling Intervals

**Default**: 5 minutes (cost-saving: 10 minutes)

**Modification**:
1. Edit `github_actions/adaptive_polling.go`
2. Modify `DefaultPollingInterval` constant
3. Rebuild and deploy

**Considerations**:
- Shorter intervals = more API calls = higher costs
- Longer intervals = delayed recording starts
- Balance between responsiveness and cost

### Custom Quality Settings

**Default**: Maximum quality (4K 60fps with fallback)

**Modification**:
1. Edit `github_actions/quality_selector.go`
2. Modify quality priority order
3. Adjust fallback logic
4. Rebuild and deploy

**Options**:
- Lock to specific quality (e.g., always 1080p60)
- Disable 4K to save bandwidth/storage
- Custom quality selection logic

### Custom Upload Targets

**Adding New Storage Provider**:

1. Edit `github_actions/storage_uploader.go`
2. Implement upload method for new provider
3. Add provider to upload sequence
4. Configure credentials as secrets
5. Test upload functionality
6. Deploy changes

**Example Providers**:
- S3-compatible storage
- Google Drive
- Dropbox
- OneDrive
- Custom HTTP endpoints

### Custom Notification Targets

**Adding New Notifier**:

1. Edit `github_actions/health_monitor.go`
2. Implement `Notifier` interface for new service
3. Add notifier to `NewHealthMonitor()`
4. Configure credentials as secrets
5. Test notification delivery
6. Deploy changes

**Example Services**:
- Slack
- Telegram
- Email (SMTP)
- SMS (Twilio)
- Custom webhooks

### Matrix Job Distribution

**Default**: Round-robin channel assignment

**Custom Distribution**:
1. Edit `github_actions/matrix_coordinator.go`
2. Modify `AssignChannels()` method
3. Implement custom assignment logic
4. Consider channel priority, load balancing, etc.
5. Deploy changes

**Strategies**:
- Priority-based assignment
- Load balancing by expected stream duration
- Geographic distribution
- Custom business logic

### Cache Optimization

**Default**: zstd compression level 19

**Tuning**:
1. Edit `.github/workflows/continuous-runner.yml`
2. Modify `ZSTD_CLEVEL` environment variable
3. Test cache save/restore times
4. Balance compression ratio vs. speed

**Compression Levels**:
- 1-3: Fast, lower compression
- 4-9: Balanced
- 10-19: Maximum compression (slower)

**Considerations**:
- Higher compression = smaller cache size
- Higher compression = longer save/restore times
- 10 GB cache limit makes high compression valuable

---

## Conclusion

You should now have a fully configured GitHub Actions Continuous Runner for GoondVR. The system will:

- ✅ Record livestreams continuously (with 30-60s gaps every 5.5 hours)
- ✅ Upload completed recordings to Gofile and Filester
- ✅ Organize recordings in a structured database
- ✅ Send notifications for all important events
- ✅ Monitor system health and disk usage
- ✅ Automatically restart before timeout
- ✅ Handle failures gracefully

### Next Steps

1. **Monitor first workflow run**: Watch logs and status file
2. **Verify uploads**: Check Gofile and Filester for completed recordings
3. **Review database**: Confirm recordings are cataloged correctly
4. **Test notifications**: Ensure alerts are received
5. **Optimize configuration**: Adjust based on your needs

### Important Reminders

- **Cost awareness**: Monitor GitHub Actions minutes usage
- **Storage management**: Regularly check Gofile/Filester storage
- **Database maintenance**: Periodically review database for completeness
- **Notification monitoring**: Respond to alerts promptly
- **Regular updates**: Keep workflow and code up to date

### Support

For issues, questions, or contributions:
- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions or share experiences
- **Pull Requests**: Contribute improvements

Happy recording! 🎬
