BEGIN;

-- Index for payload->>'action' (heavily used in WHERE clauses)
-- Used in: star events, release events, workflow events, PR filtering
CREATE INDEX IF NOT EXISTS idx_github_payload_action
    ON cauldron.github ((payload->>'action'));

-- Index for sender type filtering (User vs Bot)
-- Used in: filtering human vs automated activity
CREATE INDEX IF NOT EXISTS idx_github_payload_sender_type
    ON cauldron.github ((payload->'sender'->>'type'));

-- Index for repository name filtering
-- Used in: repo-specific queries across all event types
CREATE INDEX IF NOT EXISTS idx_github_payload_repo_name
    ON cauldron.github ((payload->'repository'->>'name'));

-- Index for organization login
-- Used in: org-level filtering and grouping
CREATE INDEX IF NOT EXISTS idx_github_payload_org_login
    ON cauldron.github ((payload->'organization'->>'login'));

-- Index for repository owner login
-- Used in: owner-based filtering when org is null
CREATE INDEX IF NOT EXISTS idx_github_payload_repo_owner_login
    ON cauldron.github ((payload->'repository'->'owner'->>'login'));

-- Index for repository owner type (Organization vs User)
-- Used in: distinguishing org repos from personal repos
CREATE INDEX IF NOT EXISTS idx_github_payload_repo_owner_type
    ON cauldron.github ((payload->'repository'->'owner'->>'type'));

-- Index for workflow run conclusion
-- Used in: workflow success/failure/cancelled filtering
CREATE INDEX IF NOT EXISTS idx_github_payload_workflow_conclusion
    ON cauldron.github ((payload->'workflow_run'->>'conclusion'));

-- Index for pull request merged flag
-- Used in: filtering merged PRs
CREATE INDEX IF NOT EXISTS idx_github_payload_pr_merged
    ON cauldron.github ((payload->'pull_request'->>'merged'));

-- Index for pull request user type
-- Used in: filtering human PRs vs bot PRs
CREATE INDEX IF NOT EXISTS idx_github_payload_pr_user_type
    ON cauldron.github ((payload->'pull_request'->'user'->>'type'));

-- Index for release prerelease flag
-- Used in: filtering stable vs pre-releases
CREATE INDEX IF NOT EXISTS idx_github_payload_release_prerelease
    ON cauldron.github ((payload->'release'->>'prerelease'));

-- Composite index for common event + action filtering pattern
-- Used in: most queries filter by event first, then action
CREATE INDEX IF NOT EXISTS idx_github_event_action
    ON cauldron.github (event, (payload->>'action'));

-- Partial composite index for workflow queries
CREATE INDEX IF NOT EXISTS idx_github_workflow_queries
    ON cauldron.github (event, (payload->>'action'), (payload->'workflow_run'->>'conclusion'))
    WHERE event = 'workflow_run';

-- Partial composite index for pull request queries
CREATE INDEX IF NOT EXISTS idx_github_pr_queries
    ON cauldron.github (event, (payload->>'action'), (payload->'pull_request'->>'merged'))
    WHERE event = 'pull_request';

-- Partial composite index for star queries
CREATE INDEX IF NOT EXISTS idx_github_star_queries
    ON cauldron.github (event, (payload->>'action'), (payload->'repository'->>'name'))
    WHERE event = 'star';

-- Partial composite index for release queries
CREATE INDEX IF NOT EXISTS idx_github_release_queries
    ON cauldron.github (event, (payload->>'action'), (payload->'repository'->>'name'))
    WHERE event = 'release';

COMMIT;
