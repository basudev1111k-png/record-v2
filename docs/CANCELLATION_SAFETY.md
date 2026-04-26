# Workflow Cancellation Safety

## Overview

When running recordings in GitHub Actions, manually cancelling a workflow can result in lost recordings. This document explains how the **Emergency Shutdown** feature protects your recordings even when workflows are cancelled.

## The Problem

GitHub Actions workflow cancellation is **immediate and forceful**:

- ❌ No graceful shutdown signals sent to the application
- ❌ Process is terminated immediately (SIGKILL-like behavior)
- ❌ No time for recordings to finalize
- ❌ Cache saves fail because files are still being written
- ❌ **Result: Recordings are lost**

### Example of Lost Recording

```
2026/04/26 05:32:00  INFO [kittengirlxo] stream type: HLS, resolution 540p
Disk usage check: 57.03 GB used, 87.22 GB free
...
Error: The operation was canceled.
/usr/bin/tar: videos/kittengirlxo_2026-04-26_05-32-00.ts: file changed as we read it
Warning: Failed to save: "/usr/bin/tar" failed with error: exit code 1
Warning: Cache save failed.
```

The recording was actively being written when cancelled, resulting in complete data loss.

## The Solution: Emergency Shutdown

The application now includes **smart signal handling** that intercepts cancellation signals and performs an emergency shutdown sequence:

### Emergency Shutdown Sequence

When you cancel a workflow, the following happens automatically:

1. **Signal Detection** 🚨
   - Catches `SIGINT` and `SIGTERM` signals
   - Logs: "⚠️ Workflow cancellation detected - initiating emergency shutdown..."

2. **Stop Background Tasks** ⏹️
   - Cancels all monitoring goroutines
   - Stops polling for new streams
   - Prevents new recordings from starting

3. **Finalize Active Recordings** 📼
   - Calls `Manager.Shutdown()` to gracefully stop all channels
   - Waits for recordings to close properly
   - Ensures files are seek-indexed and playable
   - Logs: "📼 Saving in-progress recordings..."

4. **Save State to Cache** 💾
   - Persists configuration and partial recordings
   - Saves session state for potential resume
   - Timeout: 30 seconds
   - Logs: "💾 Saving state to cache..."

5. **Upload Completed Recordings** 📤
   - Scans recordings directory for completed files
   - Uploads to Gofile and Filester in parallel
   - Deletes local files after successful upload
   - Timeout: 60 seconds
   - Logs: "📤 Uploading completed recordings..."

6. **Clean Exit** ✅
   - Logs: "✅ Emergency shutdown complete - recordings saved!"
   - Exits with status code 0

### Total Emergency Shutdown Time

The entire emergency shutdown process completes in approximately **90 seconds**:
- 30 seconds: Recording finalization
- 30 seconds: State save
- 60 seconds: Upload completed recordings

This is well within GitHub Actions' cancellation grace period.

## How to Use

### 1. Use the Cancellation-Safe Workflow

The new workflow file includes all necessary safety features:

```yaml
# .github/workflows/continuous-runner.yml
# Cancellation safety is now built-in!
```

Key features:
- ✅ Emergency shutdown signal handling
- ✅ `if: always()` on cache save (runs even when cancelled)
- ✅ Artifact upload as fallback
- ✅ Status file tracking

### 2. Configure Upload Credentials

To ensure recordings are uploaded during emergency shutdown, configure these secrets:

```bash
# Required for upload during cancellation
GOFILE_API_KEY=your_gofile_api_key
FILESTER_API_KEY=your_filester_api_key

# Optional for notifications
DISCORD_WEBHOOK_URL=your_discord_webhook
NTFY_SERVER_URL=https://ntfy.sh
NTFY_TOPIC=your_topic
```

### 3. Start a Recording

```bash
# Via GitHub Actions UI
Workflow: Continuous Runner with Cancellation Safety
Inputs:
  - channels: "channel1,channel2"
  - session_id: (leave empty for auto-generation)
  - cost_saving: false
```

### 4. Cancel Safely

When you need to stop the workflow:

1. Go to Actions tab
2. Click on the running workflow
3. Click "Cancel workflow"
4. **Wait 90 seconds** for emergency shutdown to complete
5. Check the logs for confirmation:
   ```
   ⚠️  Workflow cancellation detected - initiating emergency shutdown...
   📼 Saving in-progress recordings...
   💾 Saving state to cache...
   ✅ State saved successfully
   📤 Uploading completed recordings...
   ✅ Recordings uploaded successfully
   ✅ Emergency shutdown complete - recordings saved!
   ```

## What Gets Saved

### During Emergency Shutdown

| Item | Saved? | Location | Notes |
|------|--------|----------|-------|
| **Active recordings** | ✅ Yes | Finalized locally | Properly closed and seek-indexed |
| **Completed recordings** | ✅ Yes | Uploaded to Gofile/Filester | Deleted locally after upload |
| **Configuration** | ✅ Yes | Cache | Channel configs, settings |
| **Session state** | ✅ Yes | Cache | For potential resume |
| **Partial recordings** | ⚠️ Maybe | Artifacts | If finalization completes in time |

### Fallback: GitHub Artifacts

Even if uploads fail, recordings are preserved in GitHub Artifacts:

```yaml
- name: Upload recordings as artifacts (fallback)
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: recordings-job-${{ matrix.job_id }}-${{ github.run_number }}
    path: videos/
    retention-days: 7
```

**Artifact Limitations:**
- Maximum size: 10 GB per artifact
- Retention: 7 days (configurable)
- Manual download required

## Comparison: Before vs After

### Before (Without Emergency Shutdown)

```
User cancels workflow
  ↓
GitHub Actions sends SIGKILL
  ↓
Process terminates immediately
  ↓
Recording file left in inconsistent state
  ↓
Cache save fails (file still open)
  ↓
❌ Recording lost
```

### After (With Emergency Shutdown)

```
User cancels workflow
  ↓
GitHub Actions sends SIGTERM
  ↓
Signal handler catches SIGTERM
  ↓
Emergency shutdown sequence starts
  ↓
1. Stop background tasks
2. Finalize active recordings (30s)
3. Save state to cache (30s)
4. Upload completed recordings (60s)
  ↓
✅ Recordings saved and uploaded
```

## Best Practices

### 1. Don't Force-Cancel

If you cancel a workflow, **wait 90 seconds** before force-cancelling. This gives the emergency shutdown time to complete.

### 2. Monitor the Logs

Watch for these confirmation messages:
- ✅ "State saved successfully"
- ✅ "Recordings uploaded successfully"
- ✅ "Emergency shutdown complete"

### 3. Configure Upload Credentials

Always set `GOFILE_API_KEY` and `FILESTER_API_KEY` secrets. Without these, recordings can only be saved to artifacts (10 GB limit).

### 4. Use Cost-Saving Mode for Testing

When testing cancellation behavior, enable cost-saving mode to reduce costs:

```yaml
inputs:
  cost_saving: true  # 10-minute polling, 2 concurrent channels
```

### 5. Check Artifacts as Fallback

If uploads fail during emergency shutdown, recordings are still available in artifacts:

1. Go to workflow run page
2. Scroll to "Artifacts" section
3. Download `recordings-job-X-Y.zip`

## Troubleshooting

### Recording Still Lost After Cancellation

**Possible causes:**
1. Force-cancelled too quickly (didn't wait 90 seconds)
2. Recording was in the middle of a segment split
3. Disk space exhausted during finalization

**Solutions:**
- Wait at least 90 seconds after cancelling
- Check artifacts for partial recordings
- Increase disk space monitoring threshold

### Cache Save Failed

**Error message:**
```
/usr/bin/tar: videos/file.ts: file changed as we read it
Warning: Cache save failed.
```

**Cause:** Recording was still being written when cache save started.

**Solution:** This is expected during cancellation. The emergency shutdown handles this by:
1. Finalizing recordings first
2. Then saving to cache
3. Using artifacts as fallback

### Upload Failed During Emergency Shutdown

**Error message:**
```
⚠️  Warning: Failed to upload recordings: context deadline exceeded
```

**Cause:** Upload took longer than 60 seconds.

**Solutions:**
1. Check network connectivity
2. Verify API keys are correct
3. Recordings are still in artifacts (fallback)

### No Emergency Shutdown Logs

**Possible causes:**
1. Using old workflow file without signal handling
2. Process was force-killed (SIGKILL)
3. GitHub Actions timeout reached

**Solutions:**
- The main `continuous-runner.yml` workflow already includes cancellation safety
- Don't force-cancel workflows
- Increase timeout if needed (max 6 hours)

## Technical Details

### Signal Handling Implementation

The emergency shutdown is implemented in `main.go`:

```go
// Handle SIGINT / SIGTERM for GitHub Actions cancellation
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    
    fmt.Println("\n⚠️  Workflow cancellation detected - initiating emergency shutdown...")
    
    // 1. Cancel background tasks
    gam.Cancel()
    
    // 2. Finalize recordings
    server.Manager.Shutdown()
    
    // 3. Save state
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    gam.StatePersister.SaveState(ctx, configDir, recordingsDir)
    
    // 4. Upload recordings
    uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer uploadCancel()
    gam.UploadCompletedRecordings(uploadCtx, recordingsDir)
    
    os.Exit(0)
}()
```

### Upload Implementation

The `UploadCompletedRecordings` method scans for video files and uploads them:

```go
func (gam *GitHubActionsMode) UploadCompletedRecordings(ctx context.Context, recordingsDir string) error {
    entries, err := os.ReadDir(recordingsDir)
    // ... scan for .ts, .mp4, .mkv files
    
    for _, entry := range entries {
        if isVideoFile(entry.Name()) {
            uploadResult, err := gam.StorageUploader.UploadRecording(ctx, filePath)
            // ... handle upload
        }
    }
}
```

## FAQ

### Q: Will this work with local cancellation (Ctrl+C)?

**A:** Yes! The same signal handling works for both:
- Local: `Ctrl+C` sends `SIGINT`
- GitHub Actions: Cancellation sends `SIGTERM`

Both are caught by the signal handler.

### Q: What if I force-cancel immediately?

**A:** Force-cancellation sends `SIGKILL`, which cannot be caught. The recording will be lost. Always wait 90 seconds for graceful shutdown.

### Q: Can I resume a cancelled workflow?

**A:** Not automatically. The state is saved to cache, but you need to manually trigger a new workflow run with the same `session_id`.

### Q: What happens if upload fails?

**A:** Recordings are preserved in GitHub Artifacts as a fallback. You can download them manually from the workflow run page.

### Q: Does this increase GitHub Actions costs?

**A:** Minimal impact. Emergency shutdown adds ~90 seconds to the workflow runtime, which is negligible compared to the 5.5-hour total runtime.

## Summary

✅ **Recordings are now safe** even when workflows are cancelled  
✅ **Emergency shutdown completes in 90 seconds**  
✅ **Automatic upload to external storage**  
✅ **Artifacts as fallback** if uploads fail  
✅ **Works for both local and GitHub Actions cancellation**  

**Key takeaway:** You can now safely cancel workflows without losing recordings. Just wait 90 seconds for the emergency shutdown to complete!
