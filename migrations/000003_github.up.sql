BEGIN;

--
-- Create "github_target_type" enum
--

CREATE TYPE "cauldron"."github_target_type" AS ENUM (
    'repository',
    'organization',
    'integration',
    'business',
    'marketplace::listing',
    'sponsorslisting'
);

--
-- Create "github" table
--

CREATE TABLE "cauldron"."github" (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "delivery_id" UUID NOT NULL,
    "event" VARCHAR(128) NOT NULL,
    "target_type" "cauldron"."github_target_type" NOT NULL,
    "target_id" BIGINT NOT NULL,
    "hook_id" BIGINT NOT NULL,
    "user_login" VARCHAR(40) NOT NULL,
    "user_id" BIGINT NOT NULL,
    "kafka_offset" BIGINT NOT NULL,
    "kafka_partition" SMALLINT NOT NULL,
    "payload" JSONB NOT NULL DEFAULT '{}'::jsonb
);

--
-- Create indexes
--

CREATE INDEX "idx_github_target_type" ON "cauldron"."github" (target_type);
CREATE INDEX "idx_github_user_login" ON "cauldron"."github" (user_login);
CREATE INDEX "idx_github_user_id" ON "cauldron"."github" (user_id);
CREATE INDEX "idx_github_payload" ON "cauldron"."github" USING gin (payload);

COMMIT;
