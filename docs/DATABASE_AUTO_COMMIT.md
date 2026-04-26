# Database Auto-Commit Feature

## Overview

The workflow now automatically commits and pushes the recording database to the repository after uploads, ensuring you never lose track of your recordings even if the workflow is cancelled or times out.

---

## 🎯 When Database is Committed

The database is automatically committed and pushed in **4 scenarios**:

### 1. After Cached Uploads (Start of Workflow)
**Trigger:** When cached recordings from previous run are uploaded  
**Commit Message:** `chore: update recording database (cached uploads) [skip ci]`

```yaml
- Runs at the beginning of each workflow
- Uploads any recordings left from previous run
- Commits database entries for uploaded files
- Happens before new recording starts
```

### 2. After Emergency Cleanup (Cancellation/Timeout)
**Trigger:** When workflow is cancelled or times out  
**Commit Message:** `chore: emergency database update [skip ci]`

```yaml
- Runs when workflow is interrupted
- Commits any database entries created before interruption
- Ensures no data loss on cancellation
- Happens before cache save
```

### 3. At Workflow Completion (Normal End)
**Trigger:** When workflow completes normally (5.5 hours)  
**Commit Message:** `chore: update recording database [skip ci]`

```yaml
- Runs at the end of each 5.5-hour session
- Commits all database entries from the session
- Happens after all uploads complete
- Final commit before workflow ends
```

### 4. Manual Trigger (If Needed)
**Trigger:** Can be added to any step manually  
**Use Case:** Custom upload scripts or special scenarios

---

## 📊 Database Structure

### Directory Layout
```
database/
├── README.md           # Documentation
├── .gitkeep           # Ensures directory is tracked
├── channel1_1714159620.json
├── channel1_1714159680.json
├── channel2_1714159640.json
└── ...
```

### JSON Entry Format
```json
{
  "channel": "channelname",
  "timestamp": "2026-04-26 19-47-00",
  "gofile_url": "https://gofile.io/d/abc123",
  "filesize": 123456789,
  "uploaded_at": "2026-04-26T19:47:00Z",
  "source": "cached_recording"
}
```

### Fields Explained

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel name extracted from filename |
| `timestamp` | string | Recording timestamp (YYYY-MM-DD HH-MM-SS) |
| `gofile_url` | string | GoFile download page URL |
| `filesize` | number | File size in bytes |
| `uploaded_at` | string | ISO 8601 timestamp of upload |
| `source` | string | "cached_recording" or "live_recording" |

---

## 🔧 Implementation Details

### Git Configuration
```yaml
git config --global user.name "github-actions[bot]"
git config --global user.email "github-actions[bot]@users.noreply.github.com"
```

### Commit Process
1. Check if database directory exists
2. Add all JSON files: `git add ./database/*.json`
3. Check for changes: `git diff --staged --quiet`
4. Commit with detailed message
5. Push with retry logic (3 attempts)

### Retry Logic
```yaml
MAX_RETRIES=3
- Attempt 1: Push immediately
- Attempt 2: Wait 5s, pull --rebase, push
- Attempt 3: Wait 5s, pull --rebase, push
- If all fail: Log error, files saved in artifacts
```

### Error Handling
- All database commits use `continue-on-error: true`
- Failed pushes don't stop the workflow
- Database files are always saved in artifacts as backup
- Retry logic handles concurrent push conflicts

---

## 📝 Commit Message Format

### Normal Commit
```
chore: update recording database [skip ci]

Job: 1
Channel: channelname
Run ID: 123456789
Session: run-20260426-194700-123456789
Timestamp: 2026-04-26T19:47:00Z
```

### Cached Upload Commit
```
chore: update recording database (cached uploads) [skip ci]

Job: 1
Channel: channelname
Run ID: 123456789
Session: run-20260426-194700-123456789
Source: Cached recordings from previous run
Timestamp: 2026-04-26T19:47:00Z
```

### Emergency Commit
```
chore: emergency database update [skip ci]

Job: 1
Channel: channelname
Run ID: 123456789
Session: run-20260426-194700-123456789
Status: Emergency/Timeout
Timestamp: 2026-04-26T19:47:00Z
```

### Why `[skip ci]`?
- Prevents infinite workflow loops
- Database commits don't trigger new workflow runs
- Saves GitHub Actions minutes
- Standard practice for automated commits

---

## 🔍 Querying the Database

### Find All Recordings for a Channel
```bash
grep -r '"channel": "channelname"' database/
```

### Count Total Recordings
```bash
find database/ -name "*.json" | wc -l
```

### Get Total Size of All Recordings
```bash
jq -s 'map(.filesize) | add' database/*.json
```

### List Recent Uploads
```bash
ls -lt database/ | head -10
```

### Get All GoFile URLs
```bash
jq -r '.gofile_url' database/*.json
```

### Find Recordings by Date
```bash
grep -r '"timestamp": "2026-04-26' database/
```

### Calculate Total Storage Used
```bash
jq -s 'map(.filesize) | add | . / 1024 / 1024 / 1024' database/*.json
# Output in GB
```

---

## 🛡️ Safety Features

### 1. Concurrent Push Protection
- Uses `git pull --rebase` before retry
- Handles multiple jobs pushing simultaneously
- Automatic conflict resolution

### 2. Backup Strategy
- Database files saved in workflow artifacts
- 7-day retention for artifacts
- Can recover from failed pushes

### 3. Non-Blocking
- Uses `continue-on-error: true`
- Failed commits don't stop recording
- Workflow continues even if push fails

### 4. Idempotent
- Only commits if there are changes
- Checks `git diff --staged --quiet`
- No empty commits

---

## 📊 Monitoring

### Check Commit History
```bash
git log --grep="recording database" --oneline
```

### View Database Commits
```bash
git log --all --oneline -- database/
```

### Check Last Database Update
```bash
git log -1 --format="%ai" -- database/
```

### Count Database Entries
```bash
git ls-files database/*.json | wc -l
```

---

## 🚨 Troubleshooting

### Problem: Push Fails After 3 Retries
**Solution:**
1. Check workflow artifacts for database files
2. Manually commit from artifacts:
   ```bash
   # Download artifact
   # Extract database files
   git add database/*.json
   git commit -m "chore: manual database update"
   git push
   ```

### Problem: Merge Conflicts
**Solution:**
- Automatic: `git pull --rebase` handles most conflicts
- Manual: Database files are append-only, safe to merge

### Problem: Missing Database Entries
**Solution:**
1. Check workflow logs for upload success
2. Check artifacts for database files
3. Verify GOFILE_API_KEY is configured

### Problem: Too Many Commits
**Solution:**
- Normal behavior - one commit per upload batch
- Use `git log --oneline` to view compact history
- Consider squashing old commits periodically

---

## 📈 Benefits

### ✅ Never Lose Track of Recordings
- Database committed after every upload
- Survives workflow cancellations
- Survives workflow timeouts

### ✅ Automatic Backup
- Database in repository (permanent)
- Database in artifacts (7 days)
- Multiple recovery options

### ✅ Easy Querying
- Standard JSON format
- Use jq, grep, or any JSON tool
- Can build custom dashboards

### ✅ Audit Trail
- Git history shows all uploads
- Commit messages include metadata
- Can track when recordings were uploaded

---

## 🔄 Workflow Integration

### Before Recording Starts
```yaml
1. Restore cache
2. Upload cached recordings
3. ✅ Commit database (cached uploads)
4. Start new recording
```

### During Recording
```yaml
1. Record stream
2. Upload to GoFile
3. Save to database (JSON file)
4. Continue recording
```

### On Cancellation/Timeout
```yaml
1. Emergency cleanup
2. ✅ Commit database (emergency)
3. Save to cache
4. Upload artifacts
```

### At Workflow End
```yaml
1. Finalize recordings
2. Upload artifacts
3. ✅ Commit database (normal)
4. Update status
```

---

## 📚 Related Documentation

- `database/README.md` - Database structure and usage
- `.github/workflows/continuous-runner.yml` - Workflow implementation
- `docs/CANCELLATION_SAFETY.md` - Cancellation handling

---

## 🎉 Summary

The database auto-commit feature ensures:
- ✅ **No data loss** - Database committed after every upload
- ✅ **Survives interruptions** - Emergency commits on cancellation
- ✅ **Easy recovery** - Multiple backup locations
- ✅ **Audit trail** - Git history tracks all uploads
- ✅ **Non-blocking** - Failed commits don't stop workflow
- ✅ **Automatic** - No manual intervention needed

Your recording database is now **automatically backed up** to the repository! 🎉
