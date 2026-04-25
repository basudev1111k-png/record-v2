# Simple Channel Setup - Complete Guide

This guide shows you the **easiest way** to add channels to GoondVR using a simple JSON file.

## Why Use channels.json?

✅ **Simple** - Just edit a text file  
✅ **Fast** - Add multiple channels at once  
✅ **Portable** - Easy to backup and share  
✅ **Automatic** - Loads on startup  
✅ **Flexible** - Works alongside the Web UI  

## The 3-Step Process

### 1️⃣ Run the Setup Script

**Windows:**
```powershell
.\setup-channels.ps1
```

**macOS/Linux:**
```bash
chmod +x setup-channels.sh && ./setup-channels.sh
```

This creates `conf/channels.json` from the example file.

### 2️⃣ Edit Your Channels

Open `conf/channels.json` and replace the examples with your channels:

**Minimal (just username and site):**
```json
[
  {
    "username": "your_streamer",
    "site": "chaturbate"
  }
]
```

**Full control (all options):**
```json
[
  {
    "username": "your_streamer",
    "site": "chaturbate",
    "is_paused": false,
    "framerate": 30,
    "resolution": 1080,
    "max_duration": 0,
    "max_filesize": 0
  }
]
```

### 3️⃣ Start GoondVR

```bash
./goondvr
```

That's it! Your channels will start recording automatically.

## Real-World Examples

### Example 1: Single Channel (Simplest)

```json
[
  {
    "username": "my_favorite_streamer",
    "site": "chaturbate"
  }
]
```

### Example 2: Multiple Channels

```json
[
  {
    "username": "streamer1",
    "site": "chaturbate"
  },
  {
    "username": "streamer2",
    "site": "stripchat"
  },
  {
    "username": "streamer3",
    "site": "chaturbate"
  }
]
```

### Example 3: Different Quality Settings

```json
[
  {
    "username": "hd_streamer",
    "site": "chaturbate",
    "resolution": 1080,
    "framerate": 60
  },
  {
    "username": "mobile_streamer",
    "site": "chaturbate",
    "resolution": 720,
    "framerate": 30
  }
]
```

### Example 4: Split Large Recordings

```json
[
  {
    "username": "long_streamer",
    "site": "chaturbate",
    "max_duration": 120,
    "max_filesize": 2048
  }
]
```

This splits recordings every 120 minutes OR 2048 MB, whichever comes first.

### Example 5: Paused Channels (Add but Don't Start)

```json
[
  {
    "username": "future_streamer",
    "site": "chaturbate",
    "is_paused": true
  }
]
```

Add channels now, start recording later via the Web UI.

## Field Reference

### Required Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `username` | string | Streamer's username | `"streamer123"` |
| `site` | string | Platform: `"chaturbate"` or `"stripchat"` | `"chaturbate"` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `is_paused` | boolean | `false` | Start paused (don't record yet) |
| `framerate` | integer | `30` | FPS (30, 60, etc.) |
| `resolution` | integer | `1080` | Height in pixels (1080, 720, 480, 360) |
| `max_duration` | integer | `0` | Split every N minutes (0 = disabled) |
| `max_filesize` | integer | `0` | Split every N MB (0 = disabled) |
| `pattern` | string | (auto) | Custom filename template |

## Common Patterns

### Pattern 1: Organize by Username

```json
{
  "username": "streamer",
  "site": "chaturbate",
  "pattern": "videos/{{.Username}}/{{.Year}}-{{.Month}}-{{.Day}}_{{.Hour}}-{{.Minute}}-{{.Second}}{{if .Sequence}}_{{.Sequence}}{{end}}"
}
```

Result: `videos/streamer/2026-01-15_14-30-00.ts`

### Pattern 2: Organize by Date

```json
{
  "username": "streamer",
  "site": "chaturbate",
  "pattern": "videos/{{.Year}}/{{.Month}}/{{.Username}}_{{.Day}}_{{.Hour}}-{{.Minute}}-{{.Second}}{{if .Sequence}}_{{.Sequence}}{{end}}"
}
```

Result: `videos/2026/01/streamer_15_14-30-00.ts`

### Pattern 3: Multi-Site Organization

```json
{
  "username": "streamer",
  "site": "stripchat",
  "pattern": "videos/{{.Site}}/{{.Username}}_{{.Year}}-{{.Month}}-{{.Day}}_{{.Hour}}-{{.Minute}}-{{.Second}}{{if .Sequence}}_{{.Sequence}}{{end}}"
}
```

Result: `videos/stripchat/streamer_2026-01-15_14-30-00.ts`

## Tips & Tricks

### 💡 Tip 1: Start Small
Begin with one channel to test your setup:
```json
[
  {
    "username": "test_streamer",
    "site": "chaturbate"
  }
]
```

### 💡 Tip 2: Use Paused for Bulk Setup
Add many channels as paused, then enable them one by one:
```json
[
  {"username": "streamer1", "site": "chaturbate", "is_paused": true},
  {"username": "streamer2", "site": "chaturbate", "is_paused": true},
  {"username": "streamer3", "site": "chaturbate", "is_paused": true}
]
```

### 💡 Tip 3: Validate Your JSON
Before starting GoondVR, check your JSON syntax at https://jsonlint.com/

### 💡 Tip 4: Backup Your Config
```bash
# Backup
cp conf/channels.json conf/channels.json.backup

# Restore
cp conf/channels.json.backup conf/channels.json
```

### 💡 Tip 5: Use the Web UI for Fine-Tuning
- Start with JSON for bulk setup
- Use Web UI (http://localhost:8080) for adjustments
- Changes in Web UI are saved back to `channels.json`

## Troubleshooting

### ❌ "Channel already exists"

**Problem:** Duplicate username + site combination

**Solution:** Each channel must be unique. Remove duplicates:
```json
// ❌ BAD - duplicate
[
  {"username": "streamer", "site": "chaturbate"},
  {"username": "streamer", "site": "chaturbate"}
]

// ✅ GOOD - unique
[
  {"username": "streamer", "site": "chaturbate"},
  {"username": "other_streamer", "site": "chaturbate"}
]
```

### ❌ "Pattern conflict"

**Problem:** Two channels would write to the same file

**Solution:** Make patterns unique (include username):
```json
// ❌ BAD - same output path
[
  {"username": "streamer1", "site": "chaturbate", "pattern": "videos/recording.ts"},
  {"username": "streamer2", "site": "chaturbate", "pattern": "videos/recording.ts"}
]

// ✅ GOOD - unique paths
[
  {"username": "streamer1", "site": "chaturbate", "pattern": "videos/{{.Username}}_{{.Year}}-{{.Month}}-{{.Day}}.ts"},
  {"username": "streamer2", "site": "chaturbate", "pattern": "videos/{{.Username}}_{{.Year}}-{{.Month}}-{{.Day}}.ts"}
]
```

### ❌ JSON Syntax Error

**Common mistakes:**

```json
// ❌ Missing comma
[
  {"username": "streamer1", "site": "chaturbate"}
  {"username": "streamer2", "site": "chaturbate"}
]

// ✅ Correct
[
  {"username": "streamer1", "site": "chaturbate"},
  {"username": "streamer2", "site": "chaturbate"}
]
```

```json
// ❌ Trailing comma
[
  {"username": "streamer1", "site": "chaturbate"},
  {"username": "streamer2", "site": "chaturbate"},
]

// ✅ Correct
[
  {"username": "streamer1", "site": "chaturbate"},
  {"username": "streamer2", "site": "chaturbate"}
]
```

```json
// ❌ Unquoted strings
[
  {username: streamer1, site: chaturbate}
]

// ✅ Correct
[
  {"username": "streamer1", "site": "chaturbate"}
]
```

### ❌ File Not Found

**Problem:** `conf/channels.json` doesn't exist

**Solution:**
```bash
# Create directory
mkdir -p conf

# Copy example
cp channels.json.example conf/channels.json
```

## Advanced: Combining with Command-Line Defaults

Set defaults via command-line, override per-channel in JSON:

```bash
# Start with 720p default
./goondvr --resolution 720 --framerate 30
```

Then in `conf/channels.json`:
```json
[
  {
    "username": "streamer1",
    "site": "chaturbate"
    // Uses 720p/30fps from command-line
  },
  {
    "username": "streamer2",
    "site": "chaturbate",
    "resolution": 1080,
    "framerate": 60
    // Overrides to 1080p/60fps
  }
]
```

## Next Steps

- 📖 [CHANNELS.md](../CHANNELS.md) - Complete channel configuration reference
- 📖 [FILE_STRUCTURE.md](../FILE_STRUCTURE.md) - Understanding file locations
- 📖 [README.md](../README.md) - Full documentation
- 🌐 Web UI at http://localhost:8080 - Visual management

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────┐
│  QUICK REFERENCE: channels.json                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Location:  conf/channels.json                          │
│  Format:    JSON array                                  │
│  Loaded:    Automatically on startup                    │
│                                                          │
│  Minimal Example:                                       │
│  [                                                       │
│    {                                                     │
│      "username": "streamer",                            │
│      "site": "chaturbate"                               │
│    }                                                     │
│  ]                                                       │
│                                                          │
│  Required:  username, site                              │
│  Optional:  is_paused, framerate, resolution,           │
│             max_duration, max_filesize, pattern         │
│                                                          │
│  Sites:     "chaturbate" or "stripchat"                 │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

**You're all set! Happy recording! 🎬**
