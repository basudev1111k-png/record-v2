# GitHub Actions Continuous Runner - Limitations and Constraints

This document provides transparent information about the limitations, constraints, and expected gaps when running GoondVR on GitHub Actions. Understanding these limitations is essential for setting realistic expectations and making informed decisions about deployment.

## Table of Contents

1. [Recording Gaps During Transitions](#recording-gaps-during-transitions)
2. [GitHub Actions 6-Hour Job Limit](#github-actions-6-hour-job-limit)
3. [GitHub Actions Minutes Usage](#github-actions-minutes-usage)
4. [Cost Estimation for Free Tier](#cost-estimation-for-free-tier)
5. [Storage Limitations](#storage-limitations)
6. [Concurrent Job Limits](#concurrent-job-limits)
7. [Network and Bandwidth Constraints](#network-and-bandwidth-constraints)
8. [Platform-Specific Limitations](#platform-specific-limitations)
9. [Operational Constraints](#operational-constraints)
10. [Recommendations and Alternatives](#recommendations-and-alternatives)

---

## Recording Gaps During Transitions

### Expected Gap Duration

**30-60 seconds of recording gap occurs every 5.5 hours during workflow transitions.**

This is an **unavoidable limitation** of the auto-restart chain pattern used to work around GitHub Actions' 6-hour job timeout.

### Why Gaps Occur

The transition process involves several sequential steps:

1. **Graceful Shutdown** (5-10 seconds):
   - Active recordings are closed
   - Files are finalized and flushed to disk
   - State is prepared for persistence

2. **State Persistence** (10-20 seconds):
   - Configuration files saved to GitHub Actions cache
   - Partial recordings uploaded to cache
   - State manifest generated and verified

3. **Upload Operations** (5-15 seconds):
   - Completed recordings uploaded to Gofile and Filester
   - Local files deleted to free disk space
   - Upload status logged and verified

4. **Workflow Dispatch** (5-10 seconds):
   - GitHub API called to trigger next workflow run
   - Session state passed via workflow inputs
   - API request processed by GitHub

5. **New Workflow Startup** (10-20 seconds):
   - GitHub schedules and starts new workflow run
   - Runner provisioned and initialized
   - Repository checked out

6. **State Restoration** (5-10 seconds):
   - Cache restored from previous run
   - Configuration files loaded
   - Application initialized

**Total transition time: 40-85 seconds** (typically 30-60 seconds)

### Impact on Recording Coverage

**Daily Impact**:
- Transitions per day: ~4.4 (every 5.5 hours)
- Gap duration per transition: 30-60 seconds
- Total daily gaps: **2-4 minutes per day per channel**

**Monthly Impact**:
- Total monthly gaps: **60-120 minutes per month per channel**
- Recording coverage: **99.86% - 99.93%**

### What Happens During Gaps

**Missed Content**:
- Any livestream content occurring during the 30-60 second gap is **not recorded**
- If a stream starts during a gap, recording begins after the new workflow starts
- If a stream ends during a gap, the ending is not captured

**No Data Loss for Active Recordings**:
- Recordings in progress are properly closed before transition
- No corruption or partial file issues
- All completed segments are preserved

### Minimizing Gap Impact

**Strategies**:

1. **Monitor Transition Times**:
   - Review status file for `last_chain_transition` timestamp
   - Identify patterns in transition duration
   - Optimize configuration to reduce transition time

2. **Prioritize Critical Channels**:
   - Use matrix job assignment to prioritize important channels
   - Consider running critical channels on separate workflows
   - Stagger transition times across workflows if running multiple

3. **Accept the Limitation**:
   - 30-60 seconds every 5.5 hours is a reasonable tradeoff for continuous operation
   - Alternative platforms may be better for zero-gap requirements
   - Consider this limitation when evaluating use cases

### When Gaps Are Unacceptable

If your use case requires **zero recording gaps**, GitHub Actions is **not suitable**. Consider:

- **Self-hosted infrastructure**: Run on your own servers with no time limits
- **Cloud VMs**: AWS EC2, Google Compute Engine, Azure VMs with persistent operation
- **Container platforms**: Docker on persistent hosts, Kubernetes clusters
- **Dedicated streaming platforms**: Purpose-built recording services

---

## GitHub Actions 6-Hour Job Limit

### Hard Timeout Constraint

**GitHub Actions enforces a strict 6-hour maximum execution time for all workflow jobs.**

This is a **platform limitation** that cannot be bypassed or extended.

### How the System Works Around It

**Auto-Restart Chain Pattern**:
- Workflow runs for **5.5 hours** (330 minutes)
- At **5.4 hours**: Graceful shutdown begins
- At **5.5 hours**: New workflow triggered via GitHub API
- New workflow restores state and continues operation

**Why 5.5 Hours Instead of 6 Hours**:
- Provides 30-minute safety buffer before hard timeout
- Allows time for graceful shutdown and state persistence
- Prevents abrupt termination and data loss
- Ensures reliable chain transitions

### Risks of the 6-Hour Limit

**If Chain Fails to Trigger**:
- Workflow terminates at 6 hours
- Recording stops permanently
- Manual intervention required to restart
- State may be lost if shutdown wasn't graceful

**Mitigation**:
- Chain Manager implements retry logic (3 attempts)
- Exponential backoff for GitHub API failures
- Notifications sent if chain trigger fails
- Status file updated with chain transition status

**Recovery**:
- Manually trigger new workflow with previous session state
- State restoration from cache enables continuity
- Minimal data loss if cache is intact

### Comparison to Continuous Operation

**Traditional Deployment** (self-hosted):
- Runs indefinitely without interruptions
- No recording gaps
- No state persistence required
- No chain management overhead

**GitHub Actions Deployment**:
- Maximum 5.5-hour continuous operation
- 30-60 second gaps every 5.5 hours
- State persistence required
- Chain management complexity
- Platform constraints and limitations

**Tradeoff**: Convenience and zero infrastructure cost vs. recording gaps and complexity

---

## GitHub Actions Minutes Usage

### What Are GitHub Actions Minutes?

**GitHub Actions minutes** are the unit of measurement for workflow execution time. Each minute a workflow runs consumes one minute from your monthly quota.

**Important**: Minutes are counted **per job**, so parallel matrix jobs multiply minute consumption.

### Usage Calculation

**Single Channel (1 Matrix Job)**:
- 24/7 operation: 24 hours/day × 60 minutes/hour = **1,440 minutes/day**
- Monthly: 1,440 × 30 days = **43,200 minutes/month**

**5 Channels (5 Matrix Jobs)**:
- Each job runs independently: 43,200 minutes/month per job
- Total: 43,200 × 5 = **216,000 minutes/month**

**10 Channels (10 Matrix Jobs)**:
- Total: 43,200 × 10 = **432,000 minutes/month**

**20 Channels (20 Matrix Jobs)**:
- Total: 43,200 × 20 = **864,000 minutes/month**

### Free Tier Quota

**GitHub Free Tier**:
- **2,000 minutes/month** for public repositories
- **2,000 minutes/month** for private repositories (separate quota)

**GitHub Pro**:
- **3,000 minutes/month**

**GitHub Team**:
- **3,000 minutes/month per user**

**GitHub Enterprise**:
- **50,000 minutes/month**

### Usage vs. Free Tier

| Configuration | Monthly Minutes | Free Tier Coverage | Overage |
|--------------|----------------|-------------------|---------|
| 1 channel, 24/7 | 43,200 | 4.6% | 41,200 minutes |
| 1 channel, 8 hrs/day | 14,400 | 13.9% | 12,400 minutes |
| 1 channel, 4 hrs/day | 7,200 | 27.8% | 5,200 minutes |
| 5 channels, 24/7 | 216,000 | 0.9% | 214,000 minutes |
| 5 channels, cost-saving | 86,400 | 2.3% | 84,400 minutes |

**Conclusion**: **24/7 operation significantly exceeds the free tier** for any configuration.

### Cost-Saving Mode Impact

**Cost-Saving Mode Configuration**:
- Maximum 2 concurrent channels (2 matrix jobs)
- Polling interval: 10 minutes (vs. 5 minutes default)
- Reduced recording starts

**Usage Calculation**:
- 2 channels × 43,200 minutes/month = **86,400 minutes/month**
- Still **43x over free tier limit**

**Conclusion**: Even cost-saving mode **significantly exceeds free tier**.

### Paid Pricing

**GitHub Actions Pricing** (after free tier exhausted):
- **$0.008 per minute** for Linux runners (standard)
- **$0.016 per minute** for Windows runners
- **$0.064 per minute** for macOS runners

**Monthly Cost Examples** (Linux runners):

| Configuration | Monthly Minutes | Free Tier | Overage Minutes | Overage Cost |
|--------------|----------------|-----------|----------------|--------------|
| 1 channel, 24/7 | 43,200 | 2,000 | 41,200 | **$329.60** |
| 5 channels, 24/7 | 216,000 | 2,000 | 214,000 | **$1,712.00** |
| 10 channels, 24/7 | 432,000 | 2,000 | 430,000 | **$3,440.00** |
| 20 channels, 24/7 | 864,000 | 2,000 | 862,000 | **$6,896.00** |

**Note**: These costs are **significantly higher** than equivalent cloud VM or self-hosted solutions.

---

## Cost Estimation for Free Tier

### Realistic Free Tier Usage

To stay within the **2,000 minutes/month free tier**, you must **severely limit operation**:

**Option 1: Single Channel, Limited Hours**:
- 1 channel, 4 hours/day, 7 days/week
- 4 hours × 60 minutes × 30 days = **7,200 minutes/month**
- **Exceeds free tier by 3.6x**

**Option 2: Single Channel, Very Limited Hours**:
- 1 channel, 1 hour/day, 7 days/week
- 1 hour × 60 minutes × 30 days = **1,800 minutes/month**
- **Within free tier** ✅

**Option 3: Single Channel, Weekend Only**:
- 1 channel, 8 hours/day, 2 days/week (weekends)
- 8 hours × 60 minutes × 8 days/month = **3,840 minutes/month**
- **Exceeds free tier by 1.9x**

### Recommendations for Free Tier Users

**1. Use Scheduled Workflows**:
- Run only during peak streaming hours (e.g., 8 PM - 12 AM)
- Use GitHub Actions scheduled triggers (`cron` syntax)
- Manually start/stop workflows as needed

**2. Single Channel Focus**:
- Prioritize one most important channel
- Rotate channels weekly or monthly
- Accept that you cannot monitor multiple channels simultaneously

**3. Manual Triggering**:
- Only run workflow when you know stream is live
- Monitor stream status externally
- Trigger workflow manually when recording is needed

**4. Consider Alternatives**:
- Self-hosted runners (no minute limits, requires own infrastructure)
- Cloud free tiers (AWS, GCP, Azure, Oracle Cloud)
- Dedicated recording services
- Local recording on personal computer

### Free Tier Viability Assessment

**GitHub Actions free tier is NOT viable for**:
- ✗ 24/7 continuous recording
- ✗ Multiple channels simultaneously
- ✗ Daily recording for several hours
- ✗ Automated unattended operation

**GitHub Actions free tier MAY work for**:
- ✓ Single channel, 1-2 hours/day
- ✓ Occasional recording sessions
- ✓ Testing and development
- ✓ Proof-of-concept deployments

**Honest Assessment**: If you need reliable, continuous recording, **GitHub Actions free tier is not suitable**. Consider paid plans or alternative platforms.

---

## Storage Limitations

### GitHub Actions Cache

**Limit**: **10 GB per repository**

**Usage**:
- Configuration files: ~1-10 MB
- Partial recordings: Variable (depends on recording duration)
- State manifest: ~1 KB
- Shared configuration: ~1-10 MB

**Per Matrix Job**:
- Each job has independent cache key
- Cache keys: `state-{session_id}-{matrix_job_id}`
- 20 matrix jobs × ~500 MB average = **~10 GB total**

**Cache Eviction**:
- Least recently used (LRU) eviction when limit exceeded
- Older session caches automatically deleted
- May cause state restoration failures for old sessions

**Mitigation**:
- Use zstd compression level 19 (up to 10x reduction)
- Upload completed recordings immediately
- Clean up partial recordings after upload
- Minimize configuration file sizes

### GitHub Actions Artifacts

**Limit**: **500 MB per repository** (free tier)

**Usage**:
- Fallback storage when Gofile/Filester uploads fail
- Temporary storage for failed uploads
- Retention: 7 days (configurable, max 90 days)

**Implications**:
- Cannot rely on artifacts for long-term storage
- Must resolve upload failures within 7 days
- Artifacts consume repository storage quota

### Runner Disk Space

**Limit**: **14 GB available disk space** on GitHub-hosted runners

**Usage**:
- Operating system and tools: ~4 GB
- Repository checkout: ~100 MB - 1 GB
- Go dependencies: ~500 MB
- Active recordings: Variable (depends on stream bitrate and duration)

**Monitoring Thresholds**:
- **10 GB**: Trigger immediate uploads
- **12 GB**: Pause new recording starts
- **13 GB**: Emergency stop oldest recording

**Typical Recording Sizes**:
- 1080p60 stream: ~2-4 GB/hour
- 4K60 stream: ~8-15 GB/hour
- Multiple concurrent recordings can quickly fill disk

**Mitigation**:
- Aggressive upload strategy
- Immediate deletion after successful upload
- Disk space monitoring every 5 minutes
- Emergency actions at high usage thresholds

### External Storage Limits

**Gofile**:
- Free tier: Unlimited storage ✅
- File size: No limit ✅
- Retention: Indefinite ✅
- Bandwidth: Subject to fair use

**Filester**:
- Free tier: Unlimited storage ✅
- File size: **10 GB per file** (automatically split if larger)
- Retention: **45 days** ⚠️
- Bandwidth: Subject to fair use

**GitHub Releases**:
- File size: **2 GB per file**
- Total: Limited by repository size quota
- Retention: Indefinite ✅

**Implications**:
- Filester files expire after 45 days (download before expiration)
- Large files (>10 GB) split into chunks for Filester
- GitHub Releases not suitable for large files

---

## Concurrent Job Limits

### GitHub Actions Concurrent Job Limits

**Free Tier**:
- **20 concurrent jobs** across all repositories
- **5 concurrent jobs** per repository (macOS/Windows)
- **20 concurrent jobs** per repository (Linux)

**Paid Plans**:
- Higher limits based on plan tier
- Can request limit increases

### Impact on Multi-Channel Recording

**Matrix Job Strategy**:
- Each channel runs in independent matrix job
- Maximum 20 channels simultaneously (free tier)
- Jobs compete for concurrent job slots

**Implications**:
- Cannot record more than 20 channels simultaneously
- Other workflows in same repository share job slots
- Other repositories in same account share job slots

**Example Scenario**:
- Repository A: 15 matrix jobs (GoondVR)
- Repository B: 10 matrix jobs (other workflow)
- Total: 25 jobs requested, but only 20 allowed
- Result: 5 jobs queued, waiting for slots

### Job Queuing

**When Concurrent Limit Exceeded**:
- Excess jobs enter queue
- Jobs start as slots become available
- No guarantee of start time

**Impact on Recording**:
- Delayed recording starts
- Potential missed content
- Unpredictable behavior

**Mitigation**:
- Limit matrix job count to stay within concurrent limits
- Avoid running other workflows simultaneously
- Use dedicated repository for GoondVR
- Consider paid plan for higher limits

---

## Network and Bandwidth Constraints

### GitHub-Hosted Runner Network

**Bandwidth**: Not officially documented, but generally high-speed

**Limitations**:
- Shared infrastructure (variable performance)
- Subject to GitHub's fair use policy
- No guaranteed bandwidth
- Potential throttling for excessive use

### Upload Bandwidth

**Gofile**:
- Upload speed varies by server
- Automatic server selection for optimal performance
- Generally fast (10-50 MB/s typical)

**Filester**:
- Upload speed varies
- Generally fast (10-50 MB/s typical)

**Implications**:
- Large files (>10 GB) may take 5-20 minutes to upload
- Upload time extends transition gap
- Multiple concurrent uploads compete for bandwidth

### Download Bandwidth (Recording Streams)

**Stream Bitrates**:
- 1080p60: 5-10 Mbps
- 4K60: 20-40 Mbps

**Multiple Concurrent Streams**:
- 5 channels × 10 Mbps = 50 Mbps
- 10 channels × 10 Mbps = 100 Mbps
- 20 channels × 10 Mbps = 200 Mbps

**Potential Issues**:
- Bandwidth saturation with many concurrent streams
- Dropped frames or recording failures
- Variable stream quality

**Mitigation**:
- Limit concurrent recordings
- Use lower quality settings if bandwidth is constrained
- Monitor recording quality and adjust

---

## Platform-Specific Limitations

### GitHub API Rate Limits

**Workflow Dispatch API**:
- **1,000 requests per hour** per repository
- Used for chain transitions

**Implications**:
- Chain transitions every 5.5 hours = ~4.4 per day = ~132 per month
- Well within rate limits for single workflow
- Multiple workflows may approach limits

**If Rate Limit Exceeded**:
- Chain transition fails
- Workflow stops at 6-hour timeout
- Manual intervention required

### Git Operations

**Database Updates**:
- Each recording completion triggers git commit + push
- Multiple matrix jobs may commit simultaneously
- Potential for git conflicts

**Conflict Resolution**:
- Automatic retry with pull-merge-push
- Up to 3 retry attempts
- JSON array append is generally conflict-free

**Potential Issues**:
- High recording frequency = many commits
- Repository size grows over time
- Git history becomes large

**Mitigation**:
- Periodic repository cleanup
- Squash old commits
- Archive old database files

### Workflow Dispatch Limitations

**Input Size Limit**:
- Workflow inputs limited to **65,535 characters**
- Session state passed as JSON string

**Implications**:
- Large session state may exceed limit
- Many partial recordings increase state size
- Complex configuration increases state size

**Mitigation**:
- Minimize session state size
- Upload partial recordings before transition
- Use cache for large state data

---

## Operational Constraints

### Manual Intervention Requirements

**When Manual Intervention Is Needed**:

1. **Chain Failure**:
   - If auto-restart chain fails after all retries
   - Must manually trigger new workflow with session state

2. **Upload Failures**:
   - If both Gofile and Filester uploads fail
   - Must manually download from GitHub Artifacts
   - Must manually upload to storage

3. **Configuration Changes**:
   - Cannot change configuration during workflow run
   - Must cancel workflow and restart with new configuration

4. **Monitoring**:
   - Must periodically check status file
   - Must respond to notification alerts
   - Must verify recordings are being captured

### No Real-Time Control

**Limitations**:
- Cannot pause/resume recordings during workflow run
- Cannot add/remove channels during workflow run
- Cannot change quality settings during workflow run
- Cannot manually trigger uploads during workflow run

**Workaround**:
- Cancel workflow and restart with new configuration
- Causes recording gap during restart

### Debugging Challenges

**Limited Observability**:
- Cannot SSH into runner
- Cannot attach debugger
- Cannot inspect live state
- Must rely on logs and status file

**Log Limitations**:
- Logs only available after step completes
- No real-time log streaming
- Large logs may be truncated
- Logs deleted after retention period

### Maintenance Windows

**When Updates Are Needed**:
- Must cancel all running workflows
- Update code/configuration
- Restart workflows manually
- Causes recording gaps during maintenance

**No Rolling Updates**:
- Cannot update one matrix job at a time
- All jobs must be stopped and restarted
- Causes gaps across all channels

---

## Recommendations and Alternatives

### When GitHub Actions Is Suitable

**Good Use Cases**:
- ✓ Testing and development
- ✓ Proof-of-concept deployments
- ✓ Occasional recording (few hours per day)
- ✓ Single channel monitoring
- ✓ Learning and experimentation
- ✓ Temporary recording needs

**Acceptable Tradeoffs**:
- 30-60 second gaps every 5.5 hours
- Limited to 1-2 hours/day on free tier
- Manual monitoring and intervention
- Platform constraints and limitations

### When GitHub Actions Is NOT Suitable

**Poor Use Cases**:
- ✗ 24/7 continuous recording
- ✗ Zero-gap recording requirements
- ✗ Multiple channels (>5) simultaneously
- ✗ Production deployments
- ✗ Mission-critical recording
- ✗ Long-term unattended operation

**Unacceptable Tradeoffs**:
- High cost on paid plans
- Recording gaps during transitions
- Manual intervention requirements
- Platform limitations and constraints

### Alternative Platforms

**1. Self-Hosted Infrastructure**:
- **Pros**: No time limits, no recording gaps, full control
- **Cons**: Requires hardware, maintenance, upfront cost
- **Cost**: $0-50/month (electricity, hardware depreciation)
- **Best for**: Long-term, reliable recording

**2. Cloud VMs (AWS, GCP, Azure)**:
- **Pros**: Scalable, reliable, no time limits
- **Cons**: Ongoing cost, requires setup and management
- **Cost**: $5-50/month depending on instance size
- **Best for**: Production deployments, multiple channels

**3. Oracle Cloud Free Tier**:
- **Pros**: Always-free compute instances, no time limits
- **Cons**: Limited resources, requires account setup
- **Cost**: $0/month (free tier)
- **Best for**: Budget-conscious users, 1-5 channels

**4. Dedicated Recording Services**:
- **Pros**: Purpose-built, reliable, zero-gap recording
- **Cons**: Subscription cost, less control
- **Cost**: $10-100/month depending on service
- **Best for**: Users who want turnkey solution

**5. Local Recording (Personal Computer)**:
- **Pros**: No cloud costs, full control, no time limits
- **Cons**: Requires computer to run 24/7, electricity cost
- **Cost**: $5-20/month (electricity)
- **Best for**: Single user, personal use

### Cost Comparison

| Platform | Setup Cost | Monthly Cost (5 channels) | Recording Gaps | Maintenance |
|----------|-----------|---------------------------|----------------|-------------|
| GitHub Actions (free) | $0 | $0 (limited hours) | 30-60s every 5.5h | Low |
| GitHub Actions (paid) | $0 | $1,712 | 30-60s every 5.5h | Low |
| AWS EC2 (t3.medium) | $0 | ~$30 | None | Medium |
| Oracle Cloud (free) | $0 | $0 | None | Medium |
| Self-hosted | $200-500 | $10-20 | None | High |
| Local PC | $0 | $5-20 | None | Low |

### Final Recommendation

**For most users, GitHub Actions is NOT the optimal solution for continuous livestream recording.**

**Consider GitHub Actions only if**:
- You need temporary recording (days/weeks, not months)
- You can accept 30-60 second gaps every 5.5 hours
- You're willing to pay $300-7000/month for paid plan
- You need quick setup without infrastructure

**For reliable, cost-effective recording, consider**:
- **Oracle Cloud Free Tier**: Best free option with no time limits
- **AWS/GCP/Azure**: Best for production deployments
- **Self-hosted**: Best for long-term, full control
- **Local PC**: Best for personal use, single channel

---

## Conclusion

This document has provided transparent information about the limitations and constraints of running GoondVR on GitHub Actions. Key takeaways:

1. **Recording Gaps**: 30-60 seconds every 5.5 hours is unavoidable
2. **6-Hour Limit**: Platform constraint requiring auto-restart chain
3. **Minutes Usage**: Significantly exceeds free tier for any meaningful use
4. **Cost**: Paid plans are expensive compared to alternatives
5. **Suitability**: Best for testing, not production

**Make an informed decision** based on your specific needs, budget, and tolerance for limitations. GitHub Actions provides a clever workaround for continuous operation, but it comes with significant tradeoffs that may not be acceptable for all use cases.

For questions or clarification, please refer to the [SETUP_GUIDE.md](SETUP_GUIDE.md) or open a GitHub issue.
