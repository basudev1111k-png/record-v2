# New Feature: Simple Channel Configuration

## Overview

GoondVR now supports a **simple JSON-based channel configuration** system that makes it incredibly easy to add and manage multiple channels.

## What's New?

### 📄 Example Configuration Files

1. **`channels.json.example`** - Full example with all options
2. **`channels.json.minimal`** - Minimal example (only required fields)

### 🛠️ Setup Scripts

1. **`setup-channels.sh`** - Automated setup for macOS/Linux
2. **`setup-channels.ps1`** - Automated setup for Windows

### 📚 Comprehensive Documentation

1. **`QUICKSTART.md`** - Get started in 3 steps
2. **`CHANNELS.md`** - Complete channel configuration reference
3. **`docs/SIMPLE_SETUP.md`** - Detailed guide with examples
4. **`docs/SETUP_COMPARISON.md`** - Compare setup methods
5. **`FILE_STRUCTURE.md`** - Understanding file locations

## Quick Start

### Before (Old Way)
```bash
# Had to use Web UI or CLI for each channel
./goondvr -u streamer1 --site chaturbate
# Then manually add more via Web UI...
```

### After (New Way)
```bash
# 1. Run setup script
./setup-channels.sh

# 2. Edit conf/channels.json
[
  {"username": "streamer1", "site": "chaturbate"},
  {"username": "streamer2", "site": "stripchat"},
  {"username": "streamer3", "site": "chaturbate"}
]

# 3. Start GoondVR - all channels load automatically!
./goondvr
```

## Key Benefits

### ✅ Simplicity
- **Minimal config:** Just `username` and `site` required
- **No restart needed:** Web UI changes save automatically
- **Works together:** JSON and Web UI complement each other

### ✅ Speed
- **Bulk operations:** Add 10+ channels in seconds
- **Copy & paste:** Duplicate similar configurations
- **Templates:** Use examples as starting points

### ✅ Portability
- **Easy backup:** One file to backup (`conf/channels.json`)
- **Version control:** Track changes in git
- **Share configs:** Copy between machines

### ✅ Flexibility
- **Minimal or full:** Use minimal or detailed configs
- **Mix and match:** Different settings per channel
- **Override defaults:** Per-channel customization

## Example Configurations

### Minimal (Simplest)
```json
[
  {
    "username": "streamer",
    "site": "chaturbate"
  }
]
```

### Full Control
```json
[
  {
    "username": "streamer",
    "site": "chaturbate",
    "is_paused": false,
    "framerate": 30,
    "resolution": 1080,
    "pattern": "videos/{{.Username}}_{{.Year}}-{{.Month}}-{{.Day}}_{{.Hour}}-{{.Minute}}-{{.Second}}{{if .Sequence}}_{{.Sequence}}{{end}}",
    "max_duration": 0,
    "max_filesize": 0
  }
]
```

### Multiple Channels with Different Settings
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
    "framerate": 30,
    "max_duration": 120
  },
  {
    "username": "stripchat_user",
    "site": "stripchat",
    "resolution": 1080,
    "framerate": 30
  }
]
```

## Documentation Structure

```
goondvr/
├── QUICKSTART.md              # 3-step quick start
├── CHANNELS.md                # Complete reference
├── FILE_STRUCTURE.md          # File locations
├── README.md                  # Main documentation (updated)
│
├── channels.json.example      # Full example
├── channels.json.minimal      # Minimal example
├── setup-channels.sh          # Setup script (Unix)
├── setup-channels.ps1         # Setup script (Windows)
│
└── docs/
    ├── SIMPLE_SETUP.md        # Detailed guide
    ├── SETUP_COMPARISON.md    # Method comparison
    └── NEW_FEATURES.md        # This file
```

## Migration Guide

### From CLI Mode
**Before:**
```bash
./goondvr -u streamer --site chaturbate
```

**After:**
```bash
# Create conf/channels.json
[{"username": "streamer", "site": "chaturbate"}]

# Start without -u flag
./goondvr
```

### From Web UI Only
**Before:**
- Manually add each channel via Web UI
- No easy backup

**After:**
- Channels automatically saved to `conf/channels.json`
- Edit JSON for bulk changes
- Easy backup and restore

## Backward Compatibility

✅ **Fully backward compatible**
- Existing Web UI functionality unchanged
- CLI mode still works
- No breaking changes
- Existing configs automatically migrated

## Technical Details

### File Location
- **Path:** `./conf/channels.json`
- **Format:** JSON array of channel objects
- **Permissions:** 600 (read/write owner only)

### Loading Behavior
1. Application starts
2. Loads `conf/channels.json` if exists
3. Creates channels from config
4. Starts monitoring automatically

### Saving Behavior
1. User makes changes via Web UI
2. Changes saved to `conf/channels.json`
3. Config persists across restarts

### Validation
- Duplicate channels rejected
- Pattern conflicts detected
- JSON syntax validated
- Required fields checked

## Use Cases

### Use Case 1: Initial Setup
```bash
# New user wants to record 5 channels
./setup-channels.sh
# Edit conf/channels.json with 5 channels
./goondvr
# All 5 channels start automatically
```

### Use Case 2: Backup & Restore
```bash
# Backup
cp conf/channels.json backup/channels-2026-01-15.json

# Restore
cp backup/channels-2026-01-15.json conf/channels.json
./goondvr
```

### Use Case 3: Multi-Machine Deployment
```bash
# Machine 1: Create config
nano conf/channels.json

# Machine 2: Copy config
scp machine1:conf/channels.json conf/
./goondvr
```

### Use Case 4: Automated Deployment
```bash
# Generate config programmatically
cat > conf/channels.json << EOF
[
  {"username": "streamer1", "site": "chaturbate"},
  {"username": "streamer2", "site": "chaturbate"}
]
EOF

# Deploy with Docker
docker-compose up -d
```

## Best Practices

### ✅ Do
- Start with minimal config
- Use setup scripts for initial setup
- Backup `conf/channels.json` regularly
- Validate JSON before starting
- Use Web UI for daily management

### ❌ Don't
- Don't edit while GoondVR is running (use Web UI instead)
- Don't create duplicate channels
- Don't use conflicting patterns
- Don't forget to backup before major changes

## Troubleshooting

### Issue: Channels not loading
**Solution:** Check JSON syntax at https://jsonlint.com/

### Issue: Duplicate channel error
**Solution:** Each username+site must be unique

### Issue: Pattern conflict
**Solution:** Ensure patterns produce unique file paths

### Issue: File not found
**Solution:** Run `./setup-channels.sh` to create config

## Future Enhancements

Potential future improvements:
- YAML format support
- Config validation tool
- Import/export via Web UI
- Bulk edit via Web UI
- Config templates

## Feedback

We'd love to hear your feedback!
- [Create an issue](https://github.com/HeapOfChaos/goondvr/issues)
- [Submit a pull request](https://github.com/HeapOfChaos/goondvr/pulls)

## Credits

This feature was designed to make GoondVR more accessible and easier to use for both new and experienced users.

---

**Happy recording! 🎬**
