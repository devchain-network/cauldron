BEGIN;

DROP INDEX IF EXISTS "cauldron"."idx_gitlab_payload";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_user_id";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_user_username";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_project_path";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_project_id";
DROP INDEX IF EXISTS "cauldron"."idx_gitlab_object_kind";

DROP TABLE IF EXISTS "cauldron"."gitlab";

COMMIT;
