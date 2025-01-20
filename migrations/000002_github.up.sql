BEGIN;

CREATE TYPE github_target_type AS ENUM (
    'repository',
    'organization',
    'integration',
    'business',
    'marketplace::listing',
    'sponsorslisting'
);

CREATE TABLE github (
    "id" SERIAL PRIMARY KEY,
    "uid" UUID DEFAULT uuid_generate_v4(),
    "created_at" TIMESTAMP DEFAULT NOW() NOT NULL,
    "delivery_id" UUID NOT NULL,
    "event" VARCHAR(128) NOT NULL,
    "target" github_target_type NOT NULL,
    "target_id" BIGINT NOT NULL,
    "hook_id" BIGINT NOT NULL,
    "user_login" VARCHAR(40) NOT NULL,
    "user_id" BIGINT NOT NULL,
    "offset" BIGINT NOT NULL,
    "partition" SERIAL NOT NULL,
    "payload" JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_github_event ON github (event);
CREATE INDEX idx_github_target ON github (target);
CREATE INDEX idx_github_user_login ON github (user_login);
CREATE INDEX idx_github_user_id ON github (user_id);
CREATE INDEX idx_github_payload ON github USING gin (payload);

COMMIT;
