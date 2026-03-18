-- Index for created_at date range filtering
-- Used in: virtually every query filters by created_at >= $1 AND created_at < $2
CREATE INDEX IF NOT EXISTS "idx_github_created_at" ON "cauldron"."github" (created_at);

-- Composite index for event + date range filtering
-- Used in: most common query pattern (WHERE event = 'X' AND created_at >= $1)
CREATE INDEX IF NOT EXISTS "idx_github_event_created_at" ON "cauldron"."github" (event, created_at);
