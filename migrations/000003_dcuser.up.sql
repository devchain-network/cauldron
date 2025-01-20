BEGIN;

CREATE TYPE git_provider AS ENUM (
    'github',
    'gitlab',
    'bitbucket'
);

CREATE TABLE dcuser (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "last_seen_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "git_provider" git_provider NOT NULL,
    "git_provider_user_id" BIGINT NOT NULL,
    "git_provider_user_name" VARCHAR(40) NOT NULL,
    "git_provider_email" VARCHAR(254),
    UNIQUE ("git_provider", "git_provider_user_id")
);

CREATE TABLE github_user (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "dcuser_id" INT REFERENCES dcuser(id) ON DELETE CASCADE,
    "user_login" VARCHAR(40) NOT NULL,
    "user_id" BIGINT NOT NULL,
    "avatar_url" VARCHAR(255),
    "html_url" VARCHAR(255),
    "meta_data" JSONB NOT NULL DEFAULT '{}'::jsonb
);


CREATE INDEX idx_github_user_user_login ON github_user (user_login);
CREATE INDEX idx_github_user_user_id ON github_user (user_id);
CREATE INDEX idx_github_user_meta_data ON github_user USING gin (meta_data);


COMMIT;