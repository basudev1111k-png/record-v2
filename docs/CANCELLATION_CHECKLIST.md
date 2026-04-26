# Workflow Cancellation Safety Checklist

Use this checklist to ensure your recordings are protected when cancelling GitHub Actions workflows.

## ✅ Pre-Flight Checklist (One-Time Setup)

Before starting any recordings, complete these steps:

- [ ] **Upload credentials configured**
  - [ ] `GOFILE_API_KEY` secret added to repository
  - [ ] `FILESTER_API_KEY` secret added to repository
  - [ ] Test: Secrets visible in Settings → Secrets → Actions

- [ ] **Workflow file updated**
  - [ ] Using `.github/workflows/continuous-runner.yml` (cancellation safety built-in)
  - [ ] Verify `if: always()` on cache save step
  - [ ] Verify `if: always()` on artifact upload step

- [ ] **Application updated**
  - [ ] Latest version with emergency shutdown feature
  - [ ] Build successful: `go build -o goondvr`
  - [ ] No compilation errors

- [ ] **Optional: Notifications configured**
  - [ ] `DISCORD_WEBHOOK_URL` secret (optional)
  - [ ] `NTFY_SERVER_URL` and `NTFY_TOPIC` secrets (optional)

## ✅ Pre-Cancellation Checklist

Before cancelling a running workflow:

- [ ] **Verify recording is active**
  - [ ] Check logs for "stream detected" message
  - [ ] Disk usage increasing (indicates active recording)
  - [ ] No error messages in recent logs

- [ ] **Check upload credentials**
  - [ ] Gofile API key is valid (not expired)
  - [ ] Filester API key is valid (not expired)
  - [ ] No upload errors in logs

- [ ] **Estimate recording size**
  - [ ] Check disk usage in logs
  - [ ] Ensure < 10 GB (artifact size limit)
  - [ ] If > 10 GB, uploads are critical (artifacts won't work)

## ✅ Cancellation Procedure

Follow these steps when cancelling:

- [ ] **1. Navigate to workflow run**
  - [ ] Go to Actions tab
  - [ ] Click on the running workflow
  - [ ] Verify it's the correct workflow

- [ ] **2. Initiate cancellation**
  - [ ] Click "Cancel workflow" button
  - [ ] Confirm cancellation in dialog
  - [ ] Note the time (for 90-second wait)

- [ ] **3. Wait for emergency shutdown**
  - [ ] **Wait at least 90 seconds** ⏱️
  - [ ] Do NOT force-cancel
  - [ ] Do NOT close browser
  - [ ] Keep logs visible

- [ ] **4. Monitor logs**
  - [ ] Look for: "⚠️ Workflow cancellation detected"
  - [ ] Look for: "📼 Saving in-progress recordings..."
  - [ ] Look for: "💾 Saving state to cache..."
  - [ ] Look for: "📤 Uploading completed recordings..."

## ✅ Post-Cancellation Verification

After the workflow completes:

- [ ] **Check emergency shutdown logs**
  - [ ] "✅ State saved successfully" appears
  - [ ] "✅ Recordings uploaded successfully" appears
  - [ ] "✅ Emergency shutdown complete" appears

- [ ] **Verify uploads**
  - [ ] Gofile URL in logs (https://gofile.io/d/...)
  - [ ] Filester URL in logs (https://filester.net/...)
  - [ ] Both URLs are accessible

- [ ] **Check cache save**
  - [ ] Cache save step completed (green checkmark)
  - [ ] No "file changed as we read it" errors
  - [ ] Cache key visible in logs

- [ ] **Verify artifacts (fallback)**
  - [ ] Scroll to "Artifacts" section
  - [ ] "recordings-job-X-Y" artifact present
  - [ ] Artifact size > 0 bytes
  - [ ] Download and verify playback (if needed)

## ✅ Troubleshooting Checklist

If something went wrong:

### Recording Not Saved

- [ ] **Did you wait 90 seconds?**
  - [ ] If no: Recording may be in artifacts
  - [ ] If yes: Check logs for errors

- [ ] **Check for force-cancel**
  - [ ] Look for "SIGKILL" in logs
  - [ ] If present: Recording likely lost
  - [ ] If not: Check artifacts

- [ ] **Verify emergency shutdown ran**
  - [ ] Search logs for "Emergency shutdown"
  - [ ] If not found: Old workflow file or force-cancel
  - [ ] If found: Check next steps

### Upload Failed

- [ ] **Check API keys**
  - [ ] Gofile API key valid?
  - [ ] Filester API key valid?
  - [ ] Keys not expired?

- [ ] **Check network**
  - [ ] Look for "timeout" in logs
  - [ ] Look for "connection refused" in logs
  - [ ] GitHub Actions network issues?

- [ ] **Check file size**
  - [ ] Recording > 5 GB?
  - [ ] Upload timeout (60s) too short?
  - [ ] Consider splitting recordings

- [ ] **Fallback: Download from artifacts**
  - [ ] Go to Artifacts section
  - [ ] Download recording
  - [ ] Upload manually to storage

### Cache Save Failed

- [ ] **Check for "file changed" error**
  - [ ] This is normal during cancellation
  - [ ] Emergency shutdown handles this
  - [ ] Check artifacts for recording

- [ ] **Verify cache step ran**
  - [ ] Look for "Save state to cache" in logs
  - [ ] Check for timeout errors
  - [ ] Verify cache key in logs

### No Emergency Shutdown Logs

- [ ] **Using old workflow file?**
  - [ ] The main `continuous-runner.yml` already has cancellation safety
  - [ ] Check for emergency shutdown logs in the output

- [ ] **Force-cancelled too quickly?**
  - [ ] Wait 90 seconds next time
  - [ ] Check artifacts for partial recording

- [ ] **Process killed with SIGKILL?**
  - [ ] This bypasses signal handling
  - [ ] Recording likely lost
  - [ ] Check artifacts as last resort

## ✅ Success Criteria

Your cancellation was successful if:

- [x] Emergency shutdown logs present
- [x] "✅ State saved successfully" in logs
- [x] "✅ Recordings uploaded successfully" in logs
- [x] Gofile URL accessible
- [x] Filester URL accessible
- [x] Recording plays correctly
- [x] No errors in logs

## ✅ Quick Reference

### Minimum Wait Time
**90 seconds** after clicking "Cancel workflow"

### Expected Log Messages
```
⚠️  Workflow cancellation detected - initiating emergency shutdown...
📼 Saving in-progress recordings...
💾 Saving state to cache...
✅ State saved successfully
📤 Uploading completed recordings...
✅ Recordings uploaded successfully
✅ Emergency shutdown complete - recordings saved!
```

### Required Secrets
- `GOFILE_API_KEY` (required)
- `FILESTER_API_KEY` (required)
- `DISCORD_WEBHOOK_URL` (optional)
- `NTFY_SERVER_URL` (optional)
- `NTFY_TOPIC` (optional)

### Fallback Options
1. **Primary:** Upload to Gofile/Filester
2. **Secondary:** GitHub Actions cache
3. **Tertiary:** GitHub Actions artifacts (7-day retention)

## 📋 Print This Checklist

For easy reference, print this checklist and keep it handy when managing workflows.

---

**Need help?** See the full documentation:
- [Quick Guide](./QUICK_CANCELLATION_GUIDE.md)
- [Complete Documentation](./CANCELLATION_SAFETY.md)
- [Flow Diagram](./CANCELLATION_FLOW_DIAGRAM.md)
