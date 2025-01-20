BEGIN;

DROP INDEX IF EXISTS "idx_github_payload";
DROP INDEX IF EXISTS "idx_github_user_id";
DROP INDEX IF EXISTS "idx_github_user_login";
DROP INDEX IF EXISTS "idx_github_target_type";

DROP TABLE IF EXISTS "github";

DROP TYPE IF EXISTS "github_target_type";

COMMIT;
