BEGIN;

DROP INDEX IF EXISTS "cauldron"."idx_gitlab_user_meta_data";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_user_user_id";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_user_user_username";

DROP TABLE IF EXISTS "cauldron"."gitlab_user";

COMMIT;
