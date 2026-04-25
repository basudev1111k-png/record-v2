# Setup Methods Comparison

GoondVR offers three ways to add and manage channels. Choose the method that best fits your workflow.

## Quick Comparison

| Feature | JSON File | Web UI | CLI |
|---------|-----------|--------|-----|
| **Ease of Setup** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Bulk Operations** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐ |
| **Real-time Control** | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Visual Feedback** | ⭐ | ⭐⭐⭐⭐⭐ | ⭐ |
| **Portability** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| **Version Control** | ⭐⭐⭐⭐⭐ | ⭐ | ⭐⭐ |
| **Best For** | Initial setup | Daily management | Single channel |

## Method 1: JSON File (Recommended for Setup)

### ✅ Pros
- **Fast bulk setup** - Add 10+ channels in seconds
- **Easy to backup** - Just copy one file
- **Version control friendly** - Track changes in git
- **Portable** - Share configs between machines
- **Scriptable** - Generate configs programmatically
- **No UI needed** - Works headless

### ❌ Cons
- **Requires restart** - Changes need app restart
- **No visual feedback** - Can't see stream status while editing
- **Manual validation** - Need to check JSON syntax

### 📖 Best For
- Initial setup with multiple channels
- Bulk configuration changes
- Automated deployments
- Backup and restore
- Sharing configurations

### 🚀 Quick Start
```bash
# 1. Run setup script
./setup-channels.sh

# 2. Edit conf/channels.json
nano conf/channels.json

# 3. Start GoondVR
./goondvr
```

### 📄 Example
```json
[
  {"username": "streamer1", "site": "chaturbate"},
  {"username": "streamer2", "site": "stripchat"},
  {"username": "streamer3", "site": "chaturbate"}
]
```

**Documentation:** [CHANNELS.md](../CHANNELS.md) | [SIMPLE_SETUP.md](SIMPLE_SETUP.md)

---

## Method 2: Web UI (Recommended for Management)

### ✅ Pros
- **Visual interface** - See all channels at a glance
- **Real-time updates** - See stream status live
- **No restart needed** - Changes apply immediately
- **Easy pause/resume** - One-click control
- **Live thumbnails** - Preview streams
- **Logs visible** - Debug issues easily

### ❌ Cons
- **One at a time** - Slower for bulk operations
- **Requires browser** - Need to open http://localhost:8080
- **Less portable** - Can't easily share configs

### 📖 Best For
- Daily channel management
- Monitoring stream status
- Quick adjustments
- Troubleshooting
- Visual preference

### 🚀 Quick Start
```bash
# 1. Start GoondVR
./goondvr

# 2. Open browser
# Visit http://localhost:8080

# 3. Click "Add Channel"
# Fill in username and site
```

### 🖼️ Interface
- Dashboard with live status
- Add/remove channels
- Pause/resume recording
- View logs
- Adjust settings

**Documentation:** [README.md](../README.md#-launching-the-web-ui)

---

## Method 3: CLI (Single Channel Mode)

### ✅ Pros
- **Immediate start** - Record right away
- **No config files** - Everything in command
- **Scriptable** - Easy to automate
- **Lightweight** - No web server

### ❌ Cons
- **Single channel only** - Can't manage multiple
- **No web UI** - No visual interface
- **Less flexible** - Hard to change settings

### 📖 Best For
- Quick one-off recordings
- Testing
- Simple use cases
- Scripting/automation
- Minimal setup

### 🚀 Quick Start
```bash
./goondvr -u streamer_name --site chaturbate
```

### 📄 Example
```bash
# Basic recording
./goondvr -u yamiodymel --site chaturbate

# With custom settings
./goondvr -u yamiodymel \
  --site chaturbate \
  --resolution 720 \
  --framerate 60 \
  --max-duration 120
```

**Documentation:** [README.md](../README.md#-using-as-a-cli-tool)

---

## Recommended Workflow

### For New Users
1. **Start with JSON** - Set up your initial channels
2. **Use Web UI** - For daily management and monitoring
3. **Keep JSON as backup** - Easy restore if needed

### For Power Users
1. **JSON for bulk** - Manage large channel lists
2. **Web UI for monitoring** - Watch stream status
3. **CLI for testing** - Quick one-off recordings

### For Automation
1. **Generate JSON** - Script your channel list
2. **Deploy with Docker** - Consistent environment
3. **Monitor via API** - Programmatic access

---

## Hybrid Approach (Best of Both Worlds)

You can use **both** JSON and Web UI together:

1. **Initial setup via JSON:**
   ```json
   [
     {"username": "streamer1", "site": "chaturbate"},
     {"username": "streamer2", "site": "chaturbate"},
     {"username": "streamer3", "site": "chaturbate"}
   ]
   ```

2. **Start GoondVR:**
   ```bash
   ./goondvr
   ```

3. **Manage via Web UI:**
   - Pause/resume channels
   - Add new channels
   - Monitor status

4. **Changes persist:**
   - Web UI updates `conf/channels.json`
   - Restart loads from JSON
   - Best of both worlds!

---

## Decision Tree

```
Do you need to add multiple channels?
├─ YES → Use JSON file
│         └─ See: CHANNELS.md
│
└─ NO → Do you want visual monitoring?
        ├─ YES → Use Web UI
        │         └─ See: README.md (Web UI section)
        │
        └─ NO → Use CLI
                  └─ See: README.md (CLI section)
```

---

## Feature Matrix

| Feature | JSON | Web UI | CLI |
|---------|------|--------|-----|
| Add multiple channels | ✅ | ⚠️ One at a time | ❌ Single only |
| Remove channels | ✅ | ✅ | ❌ |
| Pause/resume | ✅ | ✅ | ❌ |
| Live status | ❌ | ✅ | ⚠️ Console only |
| Live thumbnails | ❌ | ✅ | ❌ |
| View logs | ❌ | ✅ | ⚠️ Console only |
| Backup/restore | ✅ | ⚠️ Manual export | ❌ |
| Version control | ✅ | ❌ | ⚠️ Via scripts |
| No restart needed | ❌ | ✅ | N/A |
| Works headless | ✅ | ❌ | ✅ |
| Scriptable | ✅ | ⚠️ Via API | ✅ |

---

## Migration Between Methods

### From CLI to JSON
```bash
# Old way (CLI)
./goondvr -u streamer1 --site chaturbate

# New way (JSON)
# 1. Create conf/channels.json:
[
  {"username": "streamer1", "site": "chaturbate"}
]

# 2. Start without -u flag:
./goondvr
```

### From JSON to Web UI
```bash
# 1. Start with JSON loaded
./goondvr

# 2. Open Web UI
# Visit http://localhost:8080

# 3. Manage visually
# Changes save back to JSON automatically
```

### From Web UI to JSON
```bash
# 1. Use Web UI to add channels
# 2. Channels are saved to conf/channels.json
# 3. Edit JSON directly for bulk changes
# 4. Restart to apply
```

---

## Summary

**Choose JSON if you:**
- Need to add many channels at once
- Want easy backup/restore
- Prefer text-based configuration
- Use version control
- Deploy via automation

**Choose Web UI if you:**
- Want visual monitoring
- Need real-time status updates
- Prefer point-and-click interface
- Want to see thumbnails
- Need to troubleshoot issues

**Choose CLI if you:**
- Only need one channel
- Want immediate recording
- Prefer command-line tools
- Need minimal setup
- Are testing/experimenting

**Best approach:** Start with JSON, manage with Web UI! 🎯
