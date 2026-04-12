--
-- create "gitlab_org_settings" table
--
CREATE TABLE "cauldron"."gitlab_org_settings" (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "group_name" VARCHAR(40) UNIQUE NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    "admin_user_id" BIGINT REFERENCES "cauldron"."app_user"(id),
    "repo_mode" VARCHAR(10) NOT NULL DEFAULT 'all' CHECK (repo_mode IN ('all', 'selected')),
    "selected_repos" JSONB NOT NULL DEFAULT '[]'::jsonb,
    "meta_data" JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX "idx_gitlab_org_settings_meta_data" ON "cauldron"."gitlab_org_settings" USING gin ("meta_data");
