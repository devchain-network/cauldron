BEGIN;

DROP INDEX IF EXISTS "cauldron"."idx_github_payload";
DROP INDEX IF EXISTS "cauldron"."idx_github_user_id";
DROP INDEX IF EXISTS "cauldron"."idx_github_user_login";
DROP INDEX IF EXISTS "cauldron"."idx_github_target_type";

DROP TABLE IF EXISTS "cauldron"."github";

DROP TYPE IF EXISTS "cauldron"."github_target_type";

COMMIT;
