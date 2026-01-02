BEGIN;

-- Drop partial composite indexes
DROP INDEX IF EXISTS cauldron.idx_github_release_queries;
DROP INDEX IF EXISTS cauldron.idx_github_star_queries;
DROP INDEX IF EXISTS cauldron.idx_github_pr_queries;
DROP INDEX IF EXISTS cauldron.idx_github_workflow_queries;
DROP INDEX IF EXISTS cauldron.idx_github_event_action;

-- Drop single expression indexes
DROP INDEX IF EXISTS cauldron.idx_github_payload_release_prerelease;
DROP INDEX IF EXISTS cauldron.idx_github_payload_pr_user_type;
DROP INDEX IF EXISTS cauldron.idx_github_payload_pr_merged;
DROP INDEX IF EXISTS cauldron.idx_github_payload_workflow_conclusion;
DROP INDEX IF EXISTS cauldron.idx_github_payload_repo_owner_type;
DROP INDEX IF EXISTS cauldron.idx_github_payload_repo_owner_login;
DROP INDEX IF EXISTS cauldron.idx_github_payload_org_login;
DROP INDEX IF EXISTS cauldron.idx_github_payload_repo_name;
DROP INDEX IF EXISTS cauldron.idx_github_payload_sender_type;
DROP INDEX IF EXISTS cauldron.idx_github_payload_action;

COMMIT;
