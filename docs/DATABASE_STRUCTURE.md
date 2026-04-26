# Database Structure - Simple & Organized

## Overview

The recording database is organized in a simple, intuitive structure:
- **One folder per channel**
- **One JSON file per date**
- **Automatic commits after uploads**

---

## Directory Structure

```
database/
├── README.md                      # Documentation
├── .gitkeep                       # Ensures directory is tracked
│
├── channelname1/                  # Channel folder
│   ├── 2026-04-26.json           # All recordings for this date
│   ├── 2026-04-27.json
│   └── 2026-04-28.json
│
├── channelname2/                  # Another channel
│   ├── 2026-04-26.json
│   └── 2026-04-27.json
│
└── example_channel/               # Example (for reference)
    └── 2026-04-26.json
```

---

## File Format

Each date file contains an **array** of recordings:

```json
[
  {
    "channel": "channelname",
    "timestamp": "2026-04-26 14-30-00",
    "gofile_url": "https://gofile.io/d/abc123",
    "filesize": 1234567890,
    "uploaded_at": "2026-04-26T14:35:00Z",
    "source": "live_recording"
  },
  {
    "channel": "channelname",
    "timestamp": "2026-04-26 19-45-30",
    "gofile_url": "https://gofile.io/d/def456",
    "filesize": 2345678901,
    "uploaded_at": "2026-04-26T20:10:00Z",
    "source": "live_recording"
  }
]
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `channel` | string | Channel name (extracted from filename) |
| `timestamp` | string | Recording timestamp (YYYY-MM-DD HH-MM-SS) |
| `gofile_url` | string | GoFile download page URL |
| `filesize` | number | File size in bytes |
| `uploaded_at` | string | ISO 8601 timestamp of upload |
| `source` | string | "live_recording" or "cached_recording" |

---

## How It Works

### 1. Recording Upload
```
Recording: channelname_2026-04-26_14-30-00.ts
↓
Extract: channel = "channelname"
         date = "2026-04-26"
         timestamp = "2026-04-26 14-30-00"
↓
Create/Update: database/channelname/2026-04-26.json
```

### 2. Database Update
```bash
# If file exists: Append to array
jq ". += [new_entry]" database/channelname/2026-04-26.json

# If file doesn't exist: Create new array
echo "[new_entry]" > database/channelname/2026-04-26.json
```

### 3. Git Commit
```bash
# Add all database files
git add database/*/*.json

# Commit with metadata
git commit -m "chore: update recording database [skip ci]"

# Push with retry
git push origin HEAD:main
```

---

## Benefits

### ✅ Easy to Browse
```bash
# List all channels
ls database/

# List recordings for a channel
ls database/channelname/

# View specific date
cat database/channelname/2026-04-26.json
```

### ✅ Organized by Date
- One file per day keeps things tidy
- Easy to find recordings by date
- Natural chronological order

### ✅ Simple Queries
```bash
# Count recordings for a channel
jq -s 'map(length) | add' database/channelname/*.json

# Get all URLs
jq -r '.[] | .gofile_url' database/channelname/*.json

# Calculate total size
jq -s 'map(.[].filesize) | add' database/channelname/*.json
```

### ✅ Git-Friendly
- Small, focused commits
- Easy to track changes
- Merge conflicts are rare
- Clear history per channel

---

## Automatic Commits

The database is automatically committed in 3 scenarios:

### 1. After Cached Uploads (Start of Workflow)
```yaml
Trigger: Cached recordings from previous run uploaded
Commit: "chore: update recording database (cached uploads) [skip ci]"
```

### 2. Emergency Commit (Cancellation/Timeout)
```yaml
Trigger: Workflow cancelled or times out
Commit: "chore: emergency database update [skip ci]"
```

### 3. Normal Commit (Workflow Completion)
```yaml
Trigger: Workflow completes normally (5.5 hours)
Commit: "chore: update recording database [skip ci]"
```

---

## Example Queries

### List All Channels
```bash
ls database/
```

### Count Total Recordings
```bash
find database/ -name "*.json" -exec jq 'length' {} \; | awk '{s+=$1} END {print s}'
```

### Get All URLs for a Channel
```bash
jq -r '.[] | .gofile_url' database/channelname/*.json
```

### Find Recordings by Date Range
```bash
ls database/channelname/2026-04-{26..28}.json
```

### Calculate Storage per Channel
```bash
for channel in database/*/; do
  size=$(jq -s 'map(.[].filesize) | add' "$channel"/*.json 2>/dev/null)
  echo "$(basename $channel): $(echo "scale=2; $size / 1024 / 1024 / 1024" | bc) GB"
done
```

### Get Latest Recording for a Channel
```bash
jq '.[-1]' database/channelname/$(ls database/channelname/ | tail -1)
```

---

## Maintenance

### Backup
```bash
# Clone repository to backup database
git clone <repo-url> backup
cd backup/database
```

### Export to CSV
```bash
# Export all recordings to CSV
jq -r '.[] | [.channel, .timestamp, .gofile_url, .filesize] | @csv' database/*/*.json > recordings.csv
```

### Generate Report
```bash
# Generate summary report
echo "Channel,Recordings,Total Size (GB)"
for channel in database/*/; do
  name=$(basename $channel)
  count=$(jq -s 'map(length) | add' "$channel"/*.json 2>/dev/null)
  size=$(jq -s 'map(.[].filesize) | add / 1024 / 1024 / 1024' "$channel"/*.json 2>/dev/null)
  echo "$name,$count,$size"
done
```

---

## Migration from Old Structure

If you have old database files (flat structure), migrate them:

```bash
# Create migration script
cat > migrate_database.sh << 'EOF'
#!/bin/bash

for file in database/*.json; do
  [ -f "$file" ] || continue
  
  # Extract channel and date from JSON
  channel=$(jq -r '.channel' "$file")
  timestamp=$(jq -r '.timestamp' "$file")
  date=$(echo "$timestamp" | cut -d' ' -f1)
  
  # Create channel directory
  mkdir -p "database/$channel"
  
  # Move to date-based file
  target="database/$channel/$date.json"
  if [ -f "$target" ]; then
    # Append to existing array
    jq -s '.[0] + .[1]' "$target" "$file" > "$target.tmp"
    mv "$target.tmp" "$target"
  else
    # Create new array
    jq -s '.' "$file" > "$target"
  fi
  
  # Remove old file
  rm "$file"
done

echo "Migration complete!"
EOF

chmod +x migrate_database.sh
./migrate_database.sh
```

---

## Summary

✅ **Simple Structure** - One folder per channel, one file per date  
✅ **Easy to Browse** - Intuitive organization  
✅ **Git-Friendly** - Small, focused commits  
✅ **Auto-Committed** - Never lose data  
✅ **Easy Queries** - Standard JSON tools  
✅ **Scalable** - Works with any number of channels  

Your recording database is now **organized, backed up, and easy to use**! 🎉
