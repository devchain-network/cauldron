CREATE INDEX IF NOT EXISTS "idx_gitlab_created_at" ON "cauldron"."gitlab" (created_at);
CREATE INDEX IF NOT EXISTS "idx_gitlab_event_created_at" ON "cauldron"."gitlab" (object_kind, created_at);
