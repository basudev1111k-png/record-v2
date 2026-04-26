# Recording Database

This directory contains JSON records of all uploaded recordings, organized by channel name.

## Directory Structure

```
database/
├── README.md
├── .gitkeep
├── channelname1/
│   ├── 2026-04-26.json
│   ├── 2026-04-27.json
│   └── 2026-04-28.json
├── channelname2/
│   ├── 2026-04-26.json
│   └── 2026-04-27.json
└── channelname3/
    └── 2026-04-26.json
```

## File Organization

- **Folder**: One folder per channel (e.g., `channelname1/`)
- **File**: One JSON file per date (e.g., `2026-04-26.json`)
- **Content**: Array of all recordings for that channel on that date

## JSON Format

Each date file contains an array of recordings:

```json
[
  {
    "channel": "channelname",
    "timestamp": "2026-04-26 19-47-00",
    "gofile_url": "https://gofile.io/d/abc123",
    "filesize": 123456789,
    "duration": 3600,
    "uploaded_at": "2026-04-26T19:47:00Z",
    "source": "live_recording"
  },
  {
    "channel": "channelname",
    "timestamp": "2026-04-26 22-30-15",
    "gofile_url": "https://gofile.io/d/def456",
    "filesize": 234567890,
    "duration": 5400,
    "uploaded_at": "2026-04-26T22:30:15Z",
    "source": "live_recording"
  }
]
```

## Automatic Updates

The database is automatically updated and committed to the repository:

1. **After successful upload** - When recordings are uploaded to GoFile/Filester
2. **After cached uploads** - When cached recordings from previous runs are uploaded
3. **On workflow completion** - At the end of each 5.5-hour recording session
4. **On emergency/timeout** - When workflow is cancelled or times out

## Querying the Database

### Find all recordings for a channel
```bash
ls database/channelname/
```

### Get recordings for a specific date
```bash
cat database/channelname/2026-04-26.json | jq '.'
```

### Count total recordings for a channel
```bash
jq -s 'map(length) | add' database/channelname/*.json
```

### Get all GoFile URLs for a channel
```bash
jq -r '.[] | .gofile_url' database/channelname/*.json
```

### Calculate total size for a channel
```bash
jq -s 'map(.[].filesize) | add' database/channelname/*.json
```

### Find recordings by timestamp
```bash
jq '.[] | select(.timestamp | startswith("2026-04-26 19"))' database/channelname/2026-04-26.json
```

## Backup

All database files are:
- ✅ Committed to the repository
- ✅ Included in workflow artifacts
- ✅ Automatically synced after each upload

## Notes

- One folder per channel for easy organization
- One file per date to group recordings
- Files are appended to (never overwritten)
- Failed uploads are not recorded
- The database is append-only (no deletions)
