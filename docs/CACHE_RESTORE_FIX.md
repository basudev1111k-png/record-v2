# Cache Restore Fix - Root Cause Analysis

## Problem Summary

When GitHub Actions workflows were cancelled, recordings were successfully saved to cache (79MB), but subsequent workflow runs could not find and restore these recordings. Instead, they were restoring empty caches (244 bytes) from the current run.

## Root Cause Analysis

### The Issue

The cache key strategy was causing GitHub Actions to match the **wrong cache**:

**Original Cache Save:**
```yaml
key: state-run-${{ github.run_id }}-job-${{ matrix.job.id }}
```
- Example: `state-run-24950536481-job-1` (79MB recording)

**Original Cache Restore:**
```yaml
key: state-run-${{ github.run_id }}-job-${{ matrix.job.id }}-never-match
restore-keys: |
  state-run-
  state-latest-job-${{ matrix.job.id }}
```

### Why It Failed

1. **Prefix Matching Problem**: The `restore-keys: state-run-` pattern matches **ALL** caches with that prefix
2. **Timing Issue**: GitHub Actions cache restore happens at the START of a run, but cache save happens at the END
3. **Most Recent Match**: GitHub Actions returns the **most recent** matching cache, not the largest one
4. **Current Run Pollution**: The current run's empty cache (saved at the end) was being matched instead of previous runs' recordings

### The Sequence of Events

1. **Run 1 (Cancelled)**:
   - Records video for 5-10 minutes
   - Workflow cancelled
   - Emergency cleanup runs
   - Cache saved: `state-run-24950536481-job-1` (79MB) ✅

2. **Run 2 (New)**:
   - Cache restore step runs
   - Searches for: `state-run-*`
   - Finds multiple matches:
     - `state-run-24950536481-job-1` (79MB) - **This is what we want!**
     - `state-run-24950747277-job-1` (244 bytes) - **Current run's empty cache**
   - GitHub Actions picks the most recent: `state-run-24950747277-job-1` ❌
   - Restores 244 bytes of empty cache
   - Upload step finds no recordings

## The Solution

### New Cache Key Strategy

**New Cache Save:**
```yaml
key: state-pending-upload-${{ github.run_id }}-job-${{ matrix.job.id }}
```
- Example: `state-pending-upload-24950536481-job-1`

**New Cache Restore:**
```yaml
key: state-pending-upload-${{ github.run_id }}-job-${{ matrix.job.id }}-never-match
restore-keys: |
  state-pending-upload-
```

### Why This Works

1. **Semantic Naming**: `state-pending-upload-` clearly indicates these are recordings waiting to be uploaded
2. **Excludes Current Run**: The primary key includes `-never-match` suffix, so it will NEVER match the current run's cache
3. **Finds Previous Runs**: The `restore-keys: state-pending-upload-` pattern will match ONLY previous runs with pending uploads
4. **Clear Intent**: The name makes it obvious this is a two-phase system (save → upload)

### Additional Improvements

1. **Enhanced Debugging**: Added detailed logging to show:
   - Cache hit status
   - Cache matched key
   - Current run ID
   - Directory contents and sizes
   - File counts

2. **Better Emergency Cleanup**: Shows exactly what's being saved:
   - Number of recordings
   - Total size
   - Cache key that will be used
   - List of files

3. **Upload Step Improvements**: More detailed output showing:
   - What cache was restored
   - What files were found
   - Upload progress and results

## Testing the Fix

### Expected Behavior

1. **Cancel a workflow after 5-10 minutes of recording**:
   ```
   Emergency cleanup step should show:
   ✅ Found 1 recording(s) (Total size: 79MB)
   ✅ These recordings will be saved to cache with key:
      state-pending-upload-24950536481-job-1
   ```

2. **Start a new workflow**:
   ```
   Cache restore step should show:
   Cache matched key: state-pending-upload-24950536481-job-1
   Total size: 79MB
   File count: 1
   ```

3. **Upload step should process the recording**:
   ```
   ✅ Found cached recordings to upload:
   -rw-r--r-- 1 runner docker 79M Apr 26 07:00 channelname_2026-04-26_07-00-00.ts
   
   📤 Uploading to Gofile...
   ✅ Gofile URL: https://gofile.io/d/xxxxx
   💾 Saving to database...
   ✅ Database entry created
   🗑️  Deleting local file (uploaded successfully)
   ```

### Verification Steps

1. Check the "Emergency cleanup" step logs for cache key
2. Check the "Restore job-specific state cache" step for matched key
3. Verify the matched key matches the saved key from previous run
4. Check the "Upload cached recordings" step for successful uploads
5. Verify recordings are deleted after successful upload

## Key Insights

1. **GitHub Actions Cache Behavior**: 
   - Cache restore uses **most recent** match, not largest
   - Cache keys must be carefully designed to avoid current-run pollution
   - `restore-keys` is a prefix match, not an exact match

2. **Two-Phase System Requirements**:
   - Phase 1 (Save): Must use a unique, identifiable key
   - Phase 2 (Restore): Must exclude current run from matches
   - Clear naming convention helps debugging

3. **Debugging is Critical**:
   - Always log cache keys being used
   - Show directory contents and sizes
   - Display matched cache keys
   - Make it easy to trace the flow

## Related Files

- `.github/workflows/continuous-runner.yml` - Main workflow with cache logic
- `docs/QUICK_CANCELLATION_GUIDE.md` - User guide for cancellation safety
- `docs/CANCELLATION_SAFETY.md` - Technical documentation
- `github_actions/github_actions_mode.go` - Go implementation

## References

- GitHub Actions Cache Documentation: https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows
- Cache Restore Keys: https://github.com/actions/cache/blob/main/tips-and-workarounds.md#update-a-cache
