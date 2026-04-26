# Recording Architecture Comparison

## Overview
This document compares the recording architecture between the current goondvr implementation and the vasud3v/record repository to understand how multiple channels are managed and recorded simultaneously.

---

## Current Implementation (goondvr)

### Architecture
- **Single-threaded approach**: Each channel runs in its own goroutine
- **Manager-based**: `manager.Manager` coordinates all channels
- **Channel lifecycle**: Create → Monitor → Record → Cleanup

### Key Components

#### 1. Manager (`manager/manager.go`)
```go
type Manager struct {
    Channels sync.Map  // Thread-safe map of channel instances
    SSE      *sse.Server
}
```

**Responsibilities:**
- Store and manage all channel instances using `sync.Map`
- Load/save channel configurations from `conf/channels.json`
- Coordinate channel lifecycle (create, pause, resume, stop)
- Publish SSE events for web UI updates
- Monitor disk space globally
- Track Cloudflare blocks across channels

**Key Methods:**
- `CreateChannel()` - Creates new channel and starts monitoring
- `LoadConfig()` - Loads channels from JSON and starts them with staggered delays
- `StopChannel()` - Stops monitoring and removes channel
- `PauseChannel()` / `ResumeChannel()` - Pause/resume monitoring
- `Shutdown()` - Gracefully stops all channels

#### 2. Channel (`channel/channel.go`)
```go
type Channel struct {
    CancelFunc context.CancelFunc
    LogCh      chan string
    UpdateCh   chan bool
    ThumbCh    chan bool
    done       chan struct{}
    
    IsOnline   bool
    File       *os.File
    Config     *entity.ChannelConfig
    
    // Synchronization
    fileMu     sync.RWMutex
    finalizeMu sync.Mutex
    finalizeWG sync.WaitGroup
    monitorMu  sync.Mutex
}
```

**Responsibilities:**
- Monitor a single channel for live status
- Record stream segments to disk
- Handle file rotation (max duration/filesize)
- Manage recording state and metadata
- Publish logs and updates via channels

#### 3. Recording Flow (`channel/channel_record.go`)

**Monitor Loop:**
```
1. Start monitoring with context
2. Retry loop with exponential backoff:
   - Check if channel is online
   - Fetch stream info
   - Get HLS playlist
   - Create output file
   - Watch segments (blocking)
   - Handle errors and retry
3. On context cancel: cleanup and exit
```

**Segment Handling:**
```go
func (ch *Channel) handleSegmentForMonitor(runID uint64, b []byte, duration float64) error {
    // Lock file mutex
    // Check if paused or stale run
    // Write MP4 init segment if needed
    // Write segment data
    // Update duration/filesize
    // Periodic sync every 10 segments
    // Check if should rotate file
    // Send SSE update
}
```

### Multi-Channel Strategy

#### Concurrent Execution
- Each channel runs in its own goroutine: `go ch.Resume(startSeq)`
- Staggered startup to prevent rate limiting: `time.After(time.Duration(startSeq) * time.Second)`
- Independent monitoring loops with separate contexts

#### Resource Management
- **File I/O**: Each channel has its own file handle
- **Network**: Shared HTTP client via `internal.Req`
- **Memory**: Segment buffers are per-channel
- **Disk**: Global monitoring via `Manager.diskMonitor()`

#### Synchronization
- `sync.Map` for thread-safe channel storage
- Per-channel mutexes for file operations
- Context cancellation for graceful shutdown
- WaitGroups for finalization tasks

---

## vasud3v/record Implementation

### Architecture
Very similar to current implementation with some enhancements:

#### Key Differences

1. **Enhanced Monitor State Management**
```go
type Channel struct {
    // Additional fields for monitor lifecycle
    monitorRunning          bool
    monitorRestartRequested bool
    monitorRunID            uint64
    monitorDone             chan struct{}
}
```

**Benefits:**
- Prevents race conditions during pause/resume
- Tracks monitor run IDs to reject stale segments
- Supports deferred restart when monitor is shutting down

2. **Improved Segment Handling**
```go
func (ch *Channel) handleSegmentForMonitor(runID uint64, b []byte, duration float64) error {
    ch.monitorMu.Lock()
    isCurrentRun := ch.monitorRunID == runID
    ch.monitorMu.Unlock()
    
    if !isCurrentRun {
        return retry.Unrecoverable(internal.ErrPaused)
    }
    // ... write segment
}
```

**Benefits:**
- Rejects late-arriving segments from old monitor runs
- Prevents data corruption during rapid pause/resume cycles
- Ensures file integrity

3. **Finalization Tracking**
```go
func (ch *Channel) startFinalization() {
    ch.finalizeMu.Lock()
    ch.finalizeCount++
    ch.finalizeWG.Add(1)
    ch.finalizeMu.Unlock()
}

func (ch *Channel) waitForFinalizations() int {
    // Wait for all finalization tasks
    ch.finalizeWG.Wait()
}
```

**Benefits:**
- Tracks pending finalization tasks (remux/transcode)
- Ensures clean shutdown waits for all tasks
- Prevents orphaned processes

4. **Stream End Detection**
```go
// In RecordStream:
err = playlist.WatchSegments(ctx, ...)
if err == nil || errors.Is(err, internal.ErrChannelOffline) {
    return internal.ErrStreamEnded
}
```

**Benefits:**
- Distinguishes between "stream ended" and "channel offline"
- Allows quick retry (10s) when stream ends vs full interval when offline
- Better responsiveness for channels that go live frequently

5. **Exponential Backoff for Cloudflare Blocks**
```go
if errors.Is(err, internal.ErrCloudflareBlocked) && ch.CFBlockCount > 1 {
    multiplier := 1 << (ch.CFBlockCount - 1) // 2^(n-1)
    if multiplier > 6 {
        multiplier = 6 // Cap at 30 minutes
    }
    base = base * time.Duration(multiplier)
}
```

**Benefits:**
- Reduces API pressure when repeatedly blocked
- Prevents permanent bans
- Automatic recovery when blocks clear

6. **Pattern Conflict Detection**
```go
func detectPatternConflict(conf *entity.ChannelConfig, existing []*entity.ChannelConfig) error {
    candidatePath, _ := renderPatternSample(conf)
    for _, other := range existing {
        otherPath, _ := renderPatternSample(other)
        if candidatePath == otherPath {
            return fmt.Errorf("pattern conflict")
        }
    }
}
```

**Benefits:**
- Prevents multiple channels writing to same file
- Validates patterns before starting recording
- Automatic migration from legacy patterns

---

## Comparison Summary

| Feature | Current (goondvr) | vasud3v/record |
|---------|-------------------|----------------|
| **Multi-channel** | ✅ sync.Map + goroutines | ✅ sync.Map + goroutines |
| **Staggered startup** | ✅ Sequential delays | ✅ Sequential delays |
| **Context cancellation** | ✅ Basic | ✅ Enhanced with run IDs |
| **Monitor state tracking** | ⚠️ Basic | ✅ Advanced (runID, restart queue) |
| **Stale segment rejection** | ❌ No | ✅ Yes (via runID) |
| **Finalization tracking** | ⚠️ Basic | ✅ Comprehensive (count + WaitGroup) |
| **Stream end detection** | ⚠️ Generic | ✅ Specific (quick retry) |
| **CF exponential backoff** | ❌ Fixed interval | ✅ Exponential with cap |
| **Pattern conflict check** | ❌ No | ✅ Yes (pre-validation) |
| **Periodic file sync** | ✅ Every 10 segments | ✅ Every 10 segments |
| **Disk monitoring** | ✅ Global | ✅ Global |
| **SSE updates** | ✅ Yes | ✅ Yes |

---

## Recommendations for Current Implementation

### High Priority

1. **Add Monitor Run ID Tracking**
   - Prevents stale segment writes during pause/resume
   - Critical for data integrity
   ```go
   type Channel struct {
       monitorRunID uint64
       // ...
   }
   ```

2. **Implement Pattern Conflict Detection**
   - Validate filename patterns before starting channels
   - Prevent accidental overwrites
   ```go
   func (m *Manager) CreateChannel(conf *entity.ChannelConfig) error {
       if err := detectPatternConflict(conf, m.getAllConfigs()); err != nil {
           return err
       }
       // ...
   }
   ```

3. **Add Stream End Detection**
   - Distinguish "stream ended" from "channel offline"
   - Enable quick retry (10s) for recently live channels
   ```go
   if err == nil || errors.Is(err, internal.ErrChannelOffline) {
       return internal.ErrStreamEnded
   }
   ```

### Medium Priority

4. **Enhance Finalization Tracking**
   - Track count of pending finalization tasks
   - Show progress in UI during shutdown
   ```go
   func (ch *Channel) waitForFinalizations() int {
       pending := ch.finalizeCount
       if pending > 0 {
           ch.Info("waiting for %d finalization task(s)", pending)
           ch.finalizeWG.Wait()
       }
       return pending
   }
   ```

5. **Implement CF Exponential Backoff**
   - Reduce API pressure during repeated blocks
   - Prevent permanent bans
   ```go
   if errors.Is(err, internal.ErrCloudflareBlocked) {
       multiplier := min(1 << (ch.CFBlockCount - 1), 6)
       delay = baseInterval * time.Duration(multiplier)
   }
   ```

### Low Priority

6. **Add Monitor Restart Queue**
   - Handle rapid pause/resume cycles gracefully
   - Prevent monitor goroutine leaks
   ```go
   type Channel struct {
       monitorRestartRequested bool
       monitorDone             chan struct{}
   }
   ```

---

## Implementation Notes

### Thread Safety
Both implementations use similar patterns:
- `sync.Map` for channel storage (lock-free reads)
- Per-channel mutexes for file operations
- Context cancellation for coordination
- Channels for event publishing

### Scalability
Both can handle dozens of channels simultaneously:
- Each channel: ~1 goroutine (monitor) + 1 goroutine (publisher)
- Memory: ~1-2 MB per channel (buffers + state)
- Disk I/O: Limited by disk throughput, not CPU
- Network: Limited by bandwidth, not connections

### Error Handling
Both use retry-go with custom delay functions:
- Exponential backoff for transient errors
- Fixed interval for expected offline states
- Context cancellation for graceful shutdown

---

## Conclusion

The vasud3v/record implementation demonstrates several important improvements over the current goondvr implementation, particularly in:

1. **Robustness**: Monitor run ID tracking prevents data corruption
2. **User Experience**: Stream end detection enables faster reconnection
3. **Reliability**: Pattern conflict detection prevents file overwrites
4. **Resilience**: Exponential backoff reduces ban risk

These enhancements maintain the same core architecture (goroutine-per-channel with sync.Map coordination) while adding critical safety and quality-of-life improvements.

The multi-channel recording strategy is fundamentally sound in both implementations - the key differences are in edge case handling and operational robustness.
