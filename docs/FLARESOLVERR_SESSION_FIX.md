# FlareSolverr Session Fix

## Problem

When running multiple channels in parallel via GitHub Actions matrix jobs, FlareSolverr was returning the Chaturbate homepage instead of the actual room pages. This caused all channels to show as offline even when they were live.

### Symptoms

```
[DEBUG] therealferrarii: Fetching room page via FlareSolverr: https://chaturbate.com/therealferrarii/
[DEBUG] therealferrarii: Room page received (length: 2497011)
[DEBUG] therealferrarii: Page title: Chaturbate - Free Adult Webcams, Live Sex, Free Sex Chat, Exhibitionist & Pornstar Free Cams
[DEBUG] therealferrarii: Has room_page markers: false, Has chaturbate markers: true
[INFO] therealferrarii: Channel appears to be offline (no room data found)
```

The page title shows the homepage title instead of the room page, indicating Chaturbate redirected the request.

## Root Cause

The code was passing pre-existing cookies to FlareSolverr when fetching room pages:

```go
// OLD CODE - BROKEN
cookies := internal.ParseCookies(server.Config.Cookies)
htmlBody, _, _, err := flare.GetWithCookiesAndUA(ctx, roomURL, cookies, headers)
```

When you pass cookies to FlareSolverr for a room page request, Chaturbate detects that:
1. The cookies were obtained from a different session (homepage visit)
2. The session isn't properly established for that specific room
3. Age verification wasn't completed in the context of this room

As a result, Chaturbate redirects the request to the homepage.

## Solution

**Let FlareSolverr establish its own fresh session for each room page request:**

```go
// NEW CODE - FIXED
// Pass nil for cookies to let FlareSolverr establish a fresh session
htmlBody, _, _, err := flare.GetWithCookiesAndUA(ctx, roomURL, nil, headers)
```

By passing `nil` for cookies, FlareSolverr's Chrome browser will:
1. Visit the room page with a clean slate
2. Automatically solve any Cloudflare challenges
3. Accept age verification (via `X-Requested-With: XMLHttpRequest` header)
4. Establish a proper session for that specific room
5. Return the actual room page HTML with `initialRoomDossier` data

## Why This Works

### GitHub Actions Architecture
- Each matrix job has its own isolated FlareSolverr service container
- Each FlareSolverr instance maintains its own browser sessions
- No session sharing between jobs = no conflicts

### FlareSolverr Session Management
- FlareSolverr's Chrome browser maintains session state internally
- When you don't pass cookies, it creates a fresh session
- Fresh sessions work better for room pages because:
  - Age verification is handled in the context of the room
  - Cloudflare sees a consistent browser fingerprint + session
  - No cookie/session mismatch issues

### Cookie Refresh No Longer Critical
The previous `RefreshCookiesWithFlareSolverr()` function is now less critical because:
- We're not using those cookies for room page requests
- Each FlareSolverr request establishes its own session
- The function can remain for potential future use with CycleTLS

## Testing

To verify the fix works:

1. Check that room pages are being fetched correctly:
```
[DEBUG] username: Fetching room page via FlareSolverr: https://chaturbate.com/username/
[DEBUG] username: Room page received (length: 2497011)
[DEBUG] username: Found initialRoomDossier using pattern: "window.initialRoomDossier = \""
[INFO] username: Room dossier - room_status="public", hls_source_present=true, viewers=123
[INFO] username: ✅ Stream detected via FlareSolverr! HLS URL found
```

2. Verify multiple channels record in parallel:
- All 20 matrix jobs should be able to check their assigned channels
- Each job should successfully detect when its channel is live
- No "channel is offline" false positives

## Files Changed

- `chaturbate/chaturbate.go`: Modified `fetchStreamViaFlareSolverr()` to pass `nil` for cookies

## Related Issues

- **Task 5**: Fixed FlareSolverr cookie domain (`.chaturbate.com`)
- **Task 6**: Removed global mutex (each job has its own FlareSolverr)
- **Task 8**: This fix - let FlareSolverr establish fresh sessions

## Key Insight

**Don't try to manage cookies manually when using FlareSolverr for room pages.**

FlareSolverr is designed to handle the entire browser session lifecycle. By letting it do its job without interference, we get:
- Automatic Cloudflare challenge solving
- Proper age verification
- Consistent browser fingerprinting
- No session conflicts between parallel jobs

This is the correct architecture for parallel recording in GitHub Actions.
