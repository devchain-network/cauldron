BEGIN;

--
-- Create "git_provider" enum
--

CREATE TYPE "cauldron"."git_provider" AS ENUM (
    'github',
    'gitlab',
    'bitbucket',
    'gitea'
);

--
-- Create "app_user" table
--

CREATE TABLE "cauldron"."app_user" (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "last_seen_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "git_provider" "cauldron"."git_provider" NOT NULL,
    "git_provider_user_id" BIGINT NOT NULL,
    "git_provider_user_name" VARCHAR(40) NOT NULL,
    "git_provider_email" VARCHAR(254),
    UNIQUE ("git_provider", "git_provider_user_id")
);

--
-- Create indexes
--

CREATE INDEX "idx_app_user_git_provider" ON "cauldron"."app_user" (git_provider);
CREATE INDEX "idx_app_user_git_provider_user_name" ON "cauldron"."app_user" (git_provider_user_name);

COMMIT;
