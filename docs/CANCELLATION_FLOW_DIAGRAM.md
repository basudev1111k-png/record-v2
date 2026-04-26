# Workflow Cancellation Flow Diagram

## Visual Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    USER CANCELS WORKFLOW                        │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│              GitHub Actions sends SIGTERM signal                │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│           Signal Handler in main.go catches SIGTERM             │
│                                                                 │
│   go func() {                                                   │
│       sigCh := make(chan os.Signal, 1)                          │
│       signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)    │
│       <-sigCh  // ← SIGNAL RECEIVED HERE                        │
│       // Emergency shutdown starts...                           │
│   }()                                                           │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    EMERGENCY SHUTDOWN SEQUENCE                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
        ┌─────────────────────┴─────────────────────┐
        ↓                                           ↓
┌──────────────────┐                    ┌──────────────────────┐
│  STEP 1: STOP    │                    │   STEP 2: FINALIZE   │
│ BACKGROUND TASKS │                    │     RECORDINGS       │
│                  │                    │                      │
│ gam.Cancel()     │                    │ server.Manager.      │
│                  │                    │   Shutdown()         │
│ • Stop polling   │                    │                      │
│ • Stop monitors  │                    │ • Stop channels      │
│ • Cancel context │                    │ • Close files        │
│                  │                    │ • Seek-index         │
│ Time: ~1 second  │                    │ • Wait for finalize  │
└──────────────────┘                    │                      │
                                        │ Time: ~30 seconds    │
                                        └──────────────────────┘
                              ↓
        ┌─────────────────────┴─────────────────────┐
        ↓                                           ↓
┌──────────────────┐                    ┌──────────────────────┐
│  STEP 3: SAVE    │                    │   STEP 4: UPLOAD     │
│   STATE TO       │                    │    COMPLETED         │
│     CACHE        │                    │    RECORDINGS        │
│                  │                    │                      │
│ gam.StatePersister│                   │ gam.UploadCompleted  │
│   .SaveState()   │                    │   Recordings()       │
│                  │                    │                      │
│ • Save config    │                    │ • Scan videos/       │
│ • Save session   │                    │ • Upload to Gofile   │
│ • Save state     │                    │ • Upload to Filester │
│                  │                    │ • Delete local files │
│ Timeout: 30s     │                    │                      │
└──────────────────┘                    │ Timeout: 60 seconds  │
                                        └──────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    PROCESS EXITS (os.Exit(0))                   │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│              GitHub Actions Post-Job Steps Run                  │
│                      (if: always())                             │
│                                                                 │
│  1. Save state to cache (actions/cache/save@v4)                │
│  2. Upload recordings as artifacts (actions/upload-artifact@v4) │
│  3. Upload status files (actions/upload-artifact@v4)            │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    ✅ RECORDINGS SAVED!                         │
│                                                                 │
│  • Recordings finalized and uploaded to Gofile/Filester        │
│  • State saved to cache for potential resume                   │
│  • Artifacts preserved as fallback (7-day retention)           │
│  • Total time: ~90 seconds                                     │
└─────────────────────────────────────────────────────────────────┘
```

## Timeline Breakdown

```
T+0s    │ User clicks "Cancel workflow"
        │
T+1s    │ ⚠️  Workflow cancellation detected
        │ 🛑 Background tasks stopped
        │
T+2s    │ 📼 Saving in-progress recordings...
        │ ├─ Stop channel monitoring
        │ ├─ Close recording files
        │ └─ Seek-index for playback
        │
T+30s   │ ✅ Recordings finalized
        │
T+31s   │ 💾 Saving state to cache...
        │ ├─ Save configuration
        │ ├─ Save session state
        │ └─ Save partial recordings
        │
T+60s   │ ✅ State saved successfully
        │
T+61s   │ 📤 Uploading completed recordings...
        │ ├─ Scan videos/ directory
        │ ├─ Upload to Gofile (parallel)
        │ ├─ Upload to Filester (parallel)
        │ └─ Delete local files
        │
T+90s   │ ✅ Recordings uploaded successfully
        │ ✅ Emergency shutdown complete!
        │
T+91s   │ Process exits
        │
T+92s   │ GitHub Actions post-job steps run
        │ ├─ Cache save (if: always())
        │ ├─ Artifact upload (if: always())
        │ └─ Status file upload (if: always())
        │
T+120s  │ 🎉 Workflow complete - recordings safe!
```

## Comparison: Before vs After

### ❌ Before (Without Emergency Shutdown)

```
User cancels
     ↓
SIGKILL sent
     ↓
Process terminated immediately
     ↓
Recording file left open
     ↓
Cache save fails
     ↓
❌ Recording lost
```

**Total time:** < 1 second  
**Result:** Data loss

### ✅ After (With Emergency Shutdown)

```
User cancels
     ↓
SIGTERM sent
     ↓
Signal handler catches
     ↓
Emergency shutdown (90s)
  ├─ Stop tasks (1s)
  ├─ Finalize recordings (30s)
  ├─ Save state (30s)
  └─ Upload recordings (60s)
     ↓
Process exits gracefully
     ↓
Post-job steps run
     ↓
✅ Recordings saved & uploaded
```

**Total time:** ~90 seconds  
**Result:** No data loss

## Key Components

### 1. Signal Handler (main.go)

```go
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    
    fmt.Println("⚠️  Workflow cancellation detected")
    
    // 1. Stop background tasks
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

### 2. Manager Shutdown (manager/manager.go)

```go
func (m *Manager) Shutdown() {
    m.Channels.Range(func(key, value any) bool {
        ch := value.(*channel.Channel)
        ch.Stop()  // ← Stops channel and waits for finalization
        return true
    })
    _ = m.SaveConfig()
}
```

### 3. Channel Stop (channel/channel.go)

```go
func (ch *Channel) Stop() {
    ch.Config.IsPaused = true
    ch.CancelFunc()
    ch.waitForMonitorStop()
    ch.waitForFinalizations()  // ← Waits for recordings to close
    ch.stopPublisher()
}
```

### 4. Upload Method (github_actions/github_actions_mode.go)

```go
func (gam *GitHubActionsMode) UploadCompletedRecordings(ctx context.Context, recordingsDir string) error {
    entries, err := os.ReadDir(recordingsDir)
    
    for _, entry := range entries {
        if isVideoFile(entry.Name()) {
            filePath := recordingsDir + "/" + entry.Name()
            uploadResult, err := gam.StorageUploader.UploadRecording(ctx, filePath)
            // ... handle upload
        }
    }
}
```

## Fallback Protection

```
Primary Path: Upload to Gofile/Filester
     ↓
     ├─ Success → Delete local file → ✅ Done
     │
     └─ Failure → Keep local file
                    ↓
            GitHub Actions Artifacts
                    ↓
            User downloads manually
                    ↓
            ✅ Recording preserved
```

## User Actions

### What You Should Do

1. ✅ **Click "Cancel workflow"**
2. ✅ **Wait 90 seconds** (don't force-cancel!)
3. ✅ **Check logs for confirmation**
4. ✅ **Verify uploads** (Gofile/Filester URLs in logs)
5. ✅ **Check artifacts** (if uploads failed)

### What You Should NOT Do

1. ❌ **Don't force-cancel immediately** (bypasses emergency shutdown)
2. ❌ **Don't close the browser** (you won't see confirmation logs)
3. ❌ **Don't assume it failed** (check logs and artifacts)

## Success Indicators

Look for these messages in the logs:

```
✅ State saved successfully
✅ Recordings uploaded successfully
✅ Emergency shutdown complete - recordings saved!
```

If you see all three, your recordings are safe!

## Troubleshooting

### "I don't see the emergency shutdown messages"

**Possible causes:**
- Using old workflow file without signal handling
- Force-cancelled too quickly
- Process was killed with SIGKILL

**Solution:** Use the new workflow file and wait 90 seconds.

### "Upload failed but I see the recording in artifacts"

**This is expected!** Artifacts are the fallback. Download the recording from the "Artifacts" section of the workflow run.

### "Cache save failed"

**This is normal during cancellation.** The emergency shutdown handles this by finalizing recordings first, then saving to cache. Check artifacts for the recording.

## Summary

The emergency shutdown feature ensures recordings are saved even when workflows are cancelled. The entire process takes ~90 seconds and includes:

1. 🛑 Stop background tasks (1s)
2. 📼 Finalize recordings (30s)
3. 💾 Save state to cache (30s)
4. 📤 Upload to external storage (60s)
5. ✅ Exit gracefully

**Result:** No data loss, automatic uploads, and peace of mind! 🎉
