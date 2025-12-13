# Generation Status Cleanup Job

## Overview

The `cleanup_stuck_generation` job is designed to automatically detect and fix stuck generation statuses and expired locks. This job should be run periodically (recommended: every 30 minutes) to ensure the content generation system stays healthy.

## What It Fixes

### 1. Stuck Generation Statuses
The job detects and fixes generation statuses that have been stuck in "queuing" or "processing" state for more than 3 hours. It handles three cases:

- **All tasks completed but status not marked complete**: When all tasks are accounted for (completed + failed = total_requested) but the status is still marked as "processing"
- **No active tasks but generation incomplete**: When there are no tasks in queue or processing, but the total doesn't match the requested amount
- **Invalid counter values**: When the sum of counters exceeds the total requested (caused by retry logic bugs)

### 2. Expired Locks
Removes generation locks that have been held for more than 3 hours, preventing users from being permanently blocked from generating content.

## Running the Job

### Manual Execution
```bash
cd backend
go run cmd/scheduler/main.go -job cleanup_stuck_generation
```

### Via Cron (Recommended)
Add to your crontab to run every 30 minutes:
```bash
*/30 * * * * cd /path/to/butter/backend && go run cmd/scheduler/main.go -job cleanup_stuck_generation >> /var/log/butter-cleanup.log 2>&1
```

### Via Docker/Kubernetes
Create a CronJob that runs the scheduler with the cleanup task:
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cleanup-stuck-generation
spec:
  schedule: "*/30 * * * *"  # Every 30 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: your-butter-backend-image
            command: ["./scheduler", "-job", "cleanup_stuck_generation"]
          restartPolicy: OnFailure
```

## Configuration

The cleanup job uses these hardcoded timeouts (defined in `internal/jobs/cleanup_stuck_generation.go`):

- **Stuck Status Duration**: 3 hours - Statuses older than this in "queuing" or "processing" state are considered stuck
- **Lock Expiration**: 3 hours - Locks held longer than this are considered expired

To change these values, modify the constants in the `Execute()` method:
```go
stuckDuration := 3 * time.Hour  // Change this value
lockDuration := 3 * time.Hour   // Change this value
```

## Monitoring

The job outputs log messages for all operations:
- Number of stuck statuses found and fixed
- Number of expired locks found and removed
- Details about each stuck status (ID, account, counters, etc.)
- Details about each expired lock (ID, account, age, etc.)

Monitor these logs to identify patterns in stuck generations that may indicate deeper issues.

## Related Documentation

- See `CONTENT_GENERATION_IMPLEMENTATION.md` for full system documentation
- See `CONTENT_GENERATION_REQUIREMENTS.md` for original requirements

## Troubleshooting

### Job Reports Many Stuck Statuses
If the job consistently finds many stuck statuses, this indicates:
1. Tasks are failing and not properly updating status counters
2. The repurposer service might be crashing or not processing tasks
3. Redis connectivity issues preventing task completion

**Action**: Review repurposer logs and check Redis queue health

### Job Reports Many Expired Locks
Frequent expired locks indicate:
1. Generation processes are taking longer than 3 hours (normal for large batches)
2. The CompleteGeneration method is not being called properly
3. Crashes or restarts during generation

**Action**: Review generation flow and increase lock duration if needed

### Job Fails to Execute
Common causes:
1. Database connection issues
2. Missing environment variables
3. Permissions issues

**Action**: Check logs for specific error messages and verify configuration
