BEGIN;

--
-- Create "gitlab_user" table
--

CREATE TABLE "cauldron"."gitlab_user" (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    "app_user_id" BIGINT REFERENCES "cauldron"."app_user"(id) ON DELETE NO ACTION,
    "user_username" VARCHAR(255) NOT NULL,
    "user_id" BIGINT NOT NULL,
    "name" VARCHAR(255),
    "avatar_url" VARCHAR(255),
    "email" VARCHAR(254),
    "meta_data" JSONB NOT NULL DEFAULT '{}'::jsonb
);

--
-- Create indexes
--

CREATE INDEX "idx_gitlab_user_user_username" ON "cauldron"."gitlab_user" (user_username);
CREATE INDEX "idx_gitlab_user_user_id" ON "cauldron"."gitlab_user" (user_id);
CREATE INDEX "idx_gitlab_user_meta_data" ON "cauldron"."gitlab_user" USING gin (meta_data);

COMMIT;
