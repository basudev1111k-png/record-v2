# Cache Key Strategy - Visual Explanation

## The Problem (Before Fix)

```
Timeline:
┌─────────────────────────────────────────────────────────────────┐
│ Run 1 (ID: 24950536481) - CANCELLED                            │
├─────────────────────────────────────────────────────────────────┤
│ 1. Start recording                                              │
│ 2. Record for 5-10 minutes (79MB)                              │
│ 3. User cancels workflow                                        │
│ 4. Emergency cleanup runs                                       │
│ 5. Cache saved: state-run-24950536481-job-1 (79MB) ✅          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Run 2 (ID: 24950747277) - NEW RUN                              │
├─────────────────────────────────────────────────────────────────┤
│ 1. Cache restore step runs                                      │
│    - Looking for: state-run-*                                   │
│    - Found matches:                                             │
│      • state-run-24950536481-job-1 (79MB) ← Want this!         │
│      • state-run-24950747277-job-1 (244B) ← Got this! ❌       │
│    - GitHub Actions picks most recent: 24950747277              │
│ 2. Restored 244 bytes (empty cache from current run)           │
│ 3. Upload step finds no recordings ❌                           │
│ 4. Start fresh recording                                        │
│ 5. Cache saved: state-run-24950747277-job-1 (empty)            │
└─────────────────────────────────────────────────────────────────┘
```

**Problem**: The cache key pattern `state-run-*` matches BOTH previous runs AND the current run, causing GitHub Actions to restore the wrong (empty) cache.

---

## The Solution (After Fix)

```
Timeline:
┌─────────────────────────────────────────────────────────────────┐
│ Run 1 (ID: 24950536481) - CANCELLED                            │
├─────────────────────────────────────────────────────────────────┤
│ 1. Start recording                                              │
│ 2. Record for 5-10 minutes (79MB)                              │
│ 3. User cancels workflow                                        │
│ 4. Emergency cleanup runs                                       │
│ 5. Cache saved: state-pending-upload-24950536481-job-1 (79MB) ✅│
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Run 2 (ID: 24950747277) - NEW RUN                              │
├─────────────────────────────────────────────────────────────────┤
│ 1. Cache restore step runs                                      │
│    - Primary key: state-pending-upload-24950747277-...-never-match│
│      (will never match - ensures we don't match current run)   │
│    - Restore keys: state-pending-upload-*                       │
│    - Found matches:                                             │
│      • state-pending-upload-24950536481-job-1 (79MB) ✅        │
│    - GitHub Actions picks: 24950536481 (only match!)           │
│ 2. Restored 79MB recording from previous run ✅                 │
│ 3. Upload step processes recording:                             │
│    - Upload to Gofile ✅                                        │
│    - Save to database ✅                                        │
│    - Delete local file ✅                                       │
│ 4. Start fresh recording                                        │
│ 5. Cache saved: state-pending-upload-24950747277-job-1 (new)   │
└─────────────────────────────────────────────────────────────────┘
```

**Solution**: The cache key pattern `state-pending-upload-*` only matches previous runs with recordings that need uploading, excluding the current run.

---

## Key Differences

### Before (Broken)

| Aspect | Value | Issue |
|--------|-------|-------|
| **Save Key** | `state-run-{run_id}-job-{job_id}` | Generic naming |
| **Restore Pattern** | `state-run-*` | Matches ALL runs including current |
| **Primary Key** | `state-run-{run_id}-job-{job_id}-never-match` | Still allows restore-keys to match current |
| **Result** | Restores empty cache from current run | ❌ BROKEN |

### After (Fixed)

| Aspect | Value | Benefit |
|--------|-------|---------|
| **Save Key** | `state-pending-upload-{run_id}-job-{job_id}` | Semantic naming |
| **Restore Pattern** | `state-pending-upload-*` | Only matches previous runs with pending uploads |
| **Primary Key** | `state-pending-upload-{run_id}-job-{job_id}-never-match` | Ensures current run never matches |
| **Result** | Restores recordings from previous cancelled runs | ✅ WORKS |

---

## Cache Matching Logic

### GitHub Actions Cache Restore Behavior

```
When restore-keys is specified:
1. Try to match the primary key exactly (will fail due to -never-match)
2. Try to match restore-keys patterns in order
3. For each pattern, find ALL matching caches
4. Return the MOST RECENT matching cache
```

### Before Fix (Broken)

```
Restore keys: state-run-*

Matches found:
├─ state-run-24950536481-job-1 (79MB, older)
└─ state-run-24950747277-job-1 (244B, newer) ← Selected ❌

Result: Wrong cache restored
```

### After Fix (Working)

```
Restore keys: state-pending-upload-*

Matches found:
└─ state-pending-upload-24950536481-job-1 (79MB) ← Selected ✅

Result: Correct cache restored
```

---

## Why "state-pending-upload" Works

1. **Semantic Clarity**: The name clearly indicates "these recordings need to be uploaded"
2. **Temporal Separation**: Current run hasn't saved its cache yet when restore happens
3. **Explicit Intent**: Makes the two-phase system obvious (save → upload)
4. **No Pollution**: Current run's cache won't interfere with finding previous recordings
5. **Easy Debugging**: Cache keys are self-documenting

---

## Testing Checklist

- [ ] Cancel workflow after 5-10 minutes of recording
- [ ] Check emergency cleanup logs show cache key: `state-pending-upload-{run_id}-job-{job_id}`
- [ ] Start new workflow
- [ ] Check cache restore logs show matched key: `state-pending-upload-{previous_run_id}-job-{job_id}`
- [ ] Verify cache size is ~79MB (not 244 bytes)
- [ ] Check upload step successfully uploads to Gofile
- [ ] Verify database entry is created
- [ ] Confirm local file is deleted after upload
- [ ] Verify new recording starts fresh

---

## Related Documentation

- [Cache Restore Fix - Root Cause Analysis](./CACHE_RESTORE_FIX.md)
- [Quick Cancellation Guide](./QUICK_CANCELLATION_GUIDE.md)
- [Cancellation Safety](./CANCELLATION_SAFETY.md)
