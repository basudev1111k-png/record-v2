# Quick Guide: Safe Workflow Cancellation

## TL;DR

✅ **You can now safely cancel GitHub Actions workflows without losing recordings!**

## How It Works

When you cancel a workflow:

1. 🚨 **Signal detected** - Application catches the cancellation
2. 📼 **Recordings saved** - Active recordings are finalized (30s)
3. 💾 **State cached** - Configuration saved for resume (30s)
4. 📤 **Files uploaded** - Completed recordings uploaded to Gofile/Filester (60s)
5. ✅ **Done!** - Total time: ~90 seconds

## What You Need to Do

### Before Cancelling

1. **Set up upload credentials** (one-time setup):
   ```
   Repository Settings → Secrets → Actions
   
   Add:
   - GOFILE_API_KEY
   - FILESTER_API_KEY
   ```

2. **Use the main workflow**:
   - File: `.github/workflows/continuous-runner.yml`
   - Cancellation safety is now built-in (no separate workflow needed)

### When Cancelling

1. Click "Cancel workflow" in GitHub Actions
2. **Wait 90 seconds** ⏱️ (don't force-cancel!)
3. Check logs for confirmation:
   ```
   ✅ Emergency shutdown complete - recordings saved!
   ```

## What Gets Saved

| Item | Status | Where |
|------|--------|-------|
| Active recordings | ✅ Saved | Finalized locally, then uploaded |
| Completed recordings | ✅ Uploaded | Gofile + Filester |
| Configuration | ✅ Cached | GitHub Actions cache |
| Partial recordings | ⚠️ Maybe | GitHub Artifacts (fallback) |

## Confirmation Messages

Look for these in the logs:

```
⚠️  Workflow cancellation detected - initiating emergency shutdown...
📼 Saving in-progress recordings...
💾 Saving state to cache...
✅ State saved successfully
📤 Uploading completed recordings...
✅ Recordings uploaded successfully
✅ Emergency shutdown complete - recordings saved!
```

## Troubleshooting

### "Recording still lost!"

**Did you wait 90 seconds?** Force-cancelling immediately bypasses the emergency shutdown.

**Check artifacts:** Even if upload failed, recordings are in the "Artifacts" section of the workflow run.

### "Upload failed"

**Check your API keys:** Make sure `GOFILE_API_KEY` and `FILESTER_API_KEY` are set correctly.

**Fallback:** Recordings are still in GitHub Artifacts (7-day retention).

### "Cache save failed"

**This is normal during cancellation.** The emergency shutdown handles this by finalizing recordings first, then saving to cache.

## Comparison

### ❌ Before (Old Behavior)

```
Cancel workflow → Process killed → Recording lost
```

### ✅ After (New Behavior)

```
Cancel workflow → Emergency shutdown (90s) → Recording saved & uploaded
```

## Best Practices

1. ✅ **Always wait 90 seconds** after cancelling
2. ✅ **Configure upload credentials** before starting recordings
3. ✅ **Monitor the logs** for confirmation messages
4. ✅ **Check artifacts** if uploads fail
5. ❌ **Don't force-cancel** unless absolutely necessary

## Example Workflow Run

```yaml
# Start recording
Workflow: Continuous Runner with Cancellation Safety
Inputs:
  channels: "channel1,channel2"
  
# Recording in progress...
# (30 minutes later)

# Cancel workflow
Actions → Running workflow → Cancel workflow

# Wait 90 seconds...

# Check logs:
✅ Emergency shutdown complete - recordings saved!

# Check uploads:
Gofile URL: https://gofile.io/d/abc123
Filester URL: https://filester.net/abc123

# Done! 🎉
```

## Need More Details?

See the full documentation: [CANCELLATION_SAFETY.md](./CANCELLATION_SAFETY.md)

## Summary

🎯 **Safe cancellation is now automatic**  
⏱️ **Takes 90 seconds to complete**  
📤 **Recordings uploaded automatically**  
💾 **Artifacts as fallback**  
✅ **No data loss!**
