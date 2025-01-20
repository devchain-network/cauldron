BEGIN;

--
-- Create "app_user" table
--

CREATE TABLE "cauldron"."github_user" (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "app_user_id" BIGINT REFERENCES "cauldron"."app_user"(id) ON DELETE CASCADE,
    "user_login" VARCHAR(40) NOT NULL,
    "user_id" BIGINT NOT NULL,
    "avatar_url" VARCHAR(255),
    "html_url" VARCHAR(255),
    "meta_data" JSONB NOT NULL DEFAULT '{}'::jsonb
);

--
-- Create indexes
--

CREATE INDEX "idx_github_user_user_login" ON "cauldron"."github_user" (user_login);
CREATE INDEX "idx_github_user_user_id" ON "cauldron"."github_user" (user_id);
CREATE INDEX "idx_github_user_meta_data" ON "cauldron"."github_user" USING gin (meta_data);

COMMIT;
