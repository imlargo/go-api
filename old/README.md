# Butter Backend

## Architecture

The backend consists of two main services:

### 1. API Service (`cmd/api`)
- Handles HTTP requests from frontend
- Submits tasks to Redis queue
- Queries task status and generation progress
- **Does NOT process video tasks** (read-only task queue access)

### 2. Repurposer Service (`cmd/repurposer`)
- Processes video generation tasks from Redis queue
- Runs worker pool to handle concurrent video processing
- Creates GeneratedContent records on completion
- Publishes status updates via Redis Pub/Sub

---

## Concurrent Content Generation

### Overview
The system uses a Redis-based distributed task queue to enable concurrent video generation, replacing the previous sequential generation approach.

**Key Benefits:**
- 🚀 **10x faster**: Generate 10 videos in ~60 minutes (vs 10 hours sequential)
- 🔄 **Persistent**: Tasks survive server restarts
- 📊 **Real-time status**: SSE updates for live progress tracking
- 🔒 **Race condition free**: Smart locking prevents duplicate generation

### Architecture

```
User Request → API Service → ContentService
                              ↓
                        Select all videos atomically
                              ↓
                        Submit to Redis Queue (concurrent)
                              ↓
Redis Queue → Repurposer Worker Pool → FFmpeg Processing
                              ↓
                        Create GeneratedContent
                              ↓
                        Publish status update (Redis Pub/Sub)
                              ↓
                        SSE → Frontend (real-time)
```

### Configuration

Required environment variables:
```bash
# Redis (Required for task queue)
REDIS_URL=localhost:6379

# Task Queue
TASK_WORKER_COUNT=7                  # Concurrent workers
TASK_TIMEOUT=30                      # Task timeout (minutes)
TASK_MAX_RETRIES=3                   # Retry attempts
TASK_INITIAL_RETRY_DELAY=30          # Initial delay (seconds)
TASK_MAX_RETRY_DELAY=1800            # Max delay (seconds)
```

### API Endpoints

#### Task Management
- `GET /api/v1/repurposer/tasks/:taskID` - Get task status
- `GET /api/v1/repurposer/tasks?account_id=X` - List tasks by account
- `POST /api/v1/repurposer/tasks/:taskID/retry` - Retry failed task
- `POST /api/v1/repurposer/tasks/:taskID/cancel` - Cancel task
- `GET /api/v1/repurposer/stats` - Queue statistics

#### Generation Status
- `GET /api/v2/content/generation/status?account_id=X` - Current status
- `GET /api/v2/content/generation/history?account_id=X` - History
- `GET /api/v2/content/generation/events` - SSE stream

### Database Tables

#### `repurposer_tasks`
Stores all task queue entries with full lifecycle tracking.

#### `account_generation_locks`
Prevents concurrent generation for the same account+type combination.

#### `account_generation_status`
Tracks progress of ongoing generation operations.

### For More Details
See [CONTENT_GENERATION_REQUIREMENTS.md](../CONTENT_GENERATION_REQUIREMENTS.md) for complete requirements.  
See [CONTENT_GENERATION_IMPLEMENTATION.md](../CONTENT_GENERATION_IMPLEMENTATION.md) for implementation status.

---

## Scheduled Jobs

The backend includes a job scheduler for running background tasks. Jobs are located in `internal/jobs/`.

### Running Jobs

To run a job manually:

```bash
go run cmd/scheduler/main.go -job <job_name>
```

### Available Jobs

- `track_onlyfans_links` - Tracks OnlyFans links
- `track_onlyfans_accounts` - Tracks OnlyFans accounts
- `track_accounts_analytics` - Tracks account analytics
- `track_posting_goals` - Tracks posting goals
- `track_post_analytics` - Tracks post analytics
- `auto_complete_orders` - Automatically completes marketplace orders
- `send_marketplace_message_digest` - Sends digest emails for batched marketplace messages
- `auto_generate_content` - Automatically identifies accounts scheduled for content generation


### Marketplace Message Email Batching

The `send_marketplace_message_digest` job implements email batching to prevent marketplace message notifications from being flagged as spam:

- Messages accumulate as unread count while users are away
- Job runs hourly to check for unread messages
- If 12+ hours passed since last check AND unread messages exist → send notification email
- Email notifies user about number of unread messages (without listing individual message content)
- After sending, updates the last checked timestamp

**Example cron setup:**
```bash
# Build the scheduler binary first
cd /path/to/backend && go build -o scheduler cmd/scheduler/main.go

# Run digest email job every hour
0 * * * * /path/to/backend/scheduler -job send_marketplace_message_digest
```

### Automatic Content Generation

The `auto_generate_content` job enables fully automated scheduled content generation for accounts:

- Accounts can enable auto-generation with the `auto_generate_enabled` field
- Accounts specify a generation hour (0-23) with the `auto_generate_hour` field (only full hours, no minutes)
- Job runs hourly to identify accounts scheduled for generation at the current hour
- **Automatically triggers content generation** for each matching account (generates 1 video by default)

**Account Configuration:**
Accounts can be configured via the API to enable auto-generation:

```json
{
  "auto_generate_enabled": true,
  "auto_generate_hour": 10
}
```

This would automatically trigger content generation daily at 10:00 AM (10:00, not 10:30).

**Cron setup:**
```bash
# Run auto-generation check every hour
0 * * * * /path/to/backend/scheduler -job auto_generate_content
```

**How it works:**
1. Job runs every hour (e.g., at 10:00, 11:00, 12:00, etc.)
2. Queries database for accounts with `auto_generate_enabled=true` AND `auto_generate_hour=current_hour`
3. For each matching account, initiates content generation via ContentService
4. Generates 1 video per account per run (can be customized in the future)
5. Logs results: successful generations and any errors

**Requirements:**
- Redis must be running (for task queue)
- Database must be accessible
- All environment variables must be configured (same as API service)

