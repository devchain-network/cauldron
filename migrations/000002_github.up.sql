BEGIN;
--
-- Create event_type ENUM
-- https://github.com/go-playground/webhooks/blob/master/github/github.go
--
CREATE TYPE github_event_type AS ENUM (
    'check_run',
    'check_suite',
    'commit_comment',
    'create',
    'delete',
    'dependabot_alert',
    'deploy_key',
    'deployment',
    'deployment_status',
    'fork',
    'gollum',
    'installation',
    'installation_repositories',
    'integration_installation',
    'integration_installation_repositories',
    'issue_comment',
    'issues',
    'label',
    'member',
    'membership',
    'milestone',
    'meta',
    'organization',
    'org_block',
    'page_build',
    'ping',
    'project_card',
    'project_column',
    'project',
    'public',
    'pull_request',
    'pull_request_review',
    'pull_request_review_comment',
    'push',
    'release',
    'repository',
    'repository_vulnerability_alert',
    'security_advisory',
    'star',
    'status',
    'team',
    'team_add',
    'watch',
    'workflow_dispatch',
    'workflow_job',
    'workflow_run',
    'github_app_authorization',
    'code_scanning_alert'
);

CREATE TYPE github_target_type AS ENUM (
    'repository',
    'organization',
    'integration',
    'business',
    'marketplace::listing',
    'sponsorslisting'
);

CREATE TABLE github (
    "id" UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    "created_at" TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    "delivery_id" UUID NOT NULL,
    "event" github_event_type NOT NULL,
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
