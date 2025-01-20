BEGIN;

DROP INDEX IF EXISTS "cauldron"."idx_github_user_meta_data";
DROP INDEX IF EXISTS "cauldron"."idx_github_user_user_id";
DROP INDEX IF EXISTS "cauldron"."idx_github_user_user_login";

DROP TABLE IF EXISTS "cauldron"."github_user";

COMMIT;