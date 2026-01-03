BEGIN;

--
-- Create "gitlab" table
--

CREATE TABLE "cauldron"."gitlab" (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    "event_uuid" UUID NOT NULL,
    "webhook_uuid" UUID NOT NULL,
    "object_kind" VARCHAR(64) NOT NULL,
    "project_id" BIGINT,
    "project_path" VARCHAR(255),
    "user_username" VARCHAR(255),
    "user_id" BIGINT,
    "kafka_offset" BIGINT NOT NULL,
    "kafka_partition" SMALLINT NOT NULL,
    "payload" JSONB NOT NULL DEFAULT '{}'::jsonb
);

--
-- Create indexes
--

CREATE INDEX "idx_gitlab_object_kind" ON "cauldron"."gitlab" (object_kind);
CREATE INDEX "idx_gitlab_project_id" ON "cauldron"."gitlab" (project_id) WHERE project_id IS NOT NULL;
CREATE INDEX "idx_gitlab_project_path" ON "cauldron"."gitlab" (project_path) WHERE project_path IS NOT NULL;
CREATE INDEX "idx_gitlab_user_username" ON "cauldron"."gitlab" (user_username) WHERE user_username IS NOT NULL;
CREATE INDEX "idx_gitlab_user_id" ON "cauldron"."gitlab" (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX "idx_gitlab_payload" ON "cauldron"."gitlab" USING gin (payload);

COMMIT;
