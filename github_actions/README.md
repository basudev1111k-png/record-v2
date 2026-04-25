# GitHub Actions Continuous Runner Components

This package provides components for running GoondVR continuously on GitHub Actions despite the 6-hour job timeout limitation.

## Components

### StatePersister

The `StatePersister` component handles state persistence between workflow runs using GitHub Actions cache. It saves and restores configuration files and partial recordings to enable seamless transitions.

#### Cache Compression

GitHub Actions cache automatically uses **zstd (Zstandard) compression** to reduce cache size. The workflow is configured to use compression level 19 (maximum without the `--ultra` flag) to maximize the effective use of the 10 GB cache limit.

**Compression Benefits:**
- Compression level 19 can reduce cache size by up to **10x** compared to default compression
- Enables storing more data within the 10 GB cache limit
- Faster cache uploads and downloads due to smaller file sizes
- Automatic fallback to gzip if zstd is not available

**Configuration:**
The compression level is set via the `ZSTD_CLEVEL` environment variable in the workflow YAML:

```yaml
- name: Save state to cache
  uses: actions/cache/save@v4
  env:
    ZSTD_CLEVEL: 19  # Maximum compression level (1-19)
  with:
    path: ./state
    key: state-${{ github.run_id }}
```

**Compression Levels:**
- **1-3**: Fast compression, lower ratio (default is 3)
- **4-9**: Balanced compression
- **10-19**: Maximum compression without --ultra flag
- **20-22**: Requires --ultra flag (not supported in GitHub Actions cache action)

**Note:** Higher compression levels increase CPU usage during cache save/restore operations but significantly reduce storage requirements, which is beneficial for maximizing the 10 GB cache limit.

#### Cache Restoration Error Handling

When restoring state from cache, the `RestoreState()` method may return different types of errors:

1. **Cache Miss** (`ErrCacheMiss`): No cached state exists. This is **expected** for the first workflow run.
2. **Integrity Failures**: Cached files have mismatched checksums or sizes.
3. **I/O Errors**: File system or permission errors.

**Proper error handling pattern:**

```go
import (
    "context"
    "errors"
    "log"
    
    "github.com/HeapOfChaos/goondvr/github_actions"
)

func restoreOrInitialize(sp *github_actions.StatePersister, configDir, recordingsDir string) error {
    ctx := context.Background()
    
    err := sp.RestoreState(ctx, configDir, recordingsDir)
    
    if errors.Is(err, github_actions.ErrCacheMiss) {
        // Cache miss is expected for first run - initialize with defaults
        log.Println("No cached state found, initializing with default configuration")
        return initializeDefaultConfiguration(configDir)
    } else if err != nil {
        // Other errors (integrity failures, I/O errors) - log warning and continue
        log.Printf("Warning: cache restoration failed: %v", err)
        log.Println("Continuing with default configuration")
        return initializeDefaultConfiguration(configDir)
    }
    
    // Cache restored successfully
    log.Println("Successfully restored state from cache")
    return nil
}

func initializeDefaultConfiguration(configDir string) error {
    // Create default configuration
    // This should match your application's default settings
    log.Println("Creating default configuration...")
    
    // Example: Create config directory and default files
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return fmt.Errorf("failed to create config directory: %w", err)
    }
    
    // Add your default configuration initialization here
    // For example, create default config.json, settings.json, etc.
    
    return nil
}
```

#### Helper Functions

The package provides helper functions to make error handling easier:

- `IsCacheMiss(err error) bool`: Returns true if the error is a cache miss error

```go
err := sp.RestoreState(ctx, configDir, recordingsDir)
if github_actions.IsCacheMiss(err) {
    // Handle cache miss
}
```

### ChainManager

The `ChainManager` component handles the auto-restart chain pattern, triggering new workflow runs before the 6-hour timeout.

(Documentation for other components will be added as they are implemented)

## Requirements Mapping

- **Requirement 2.7**: Cache restoration failures are handled gracefully by returning `ErrCacheMiss` for missing cache entries, allowing callers to initialize with default configuration and continue operation.
- **Requirement 2.1-2.6, 2.8**: State persistence and restoration with integrity verification.
- **Requirement 9.5**: GitHub Actions cache compression using zstd level 19 to maximize the 10 GB cache limit.

## Testing

Run tests with:

```bash
go test ./github_actions/...
```

Run tests with coverage:

```bash
go test -cover ./github_actions/...
```
