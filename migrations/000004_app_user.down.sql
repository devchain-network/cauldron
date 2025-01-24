BEGIN;

DROP INDEX IF EXISTS "cauldron"."idx_app_user_git_provider_user_name";
DROP INDEX IF EXISTS "cauldron"."idx_app_user_git_provider";

DROP TABLE IF EXISTS "cauldron"."app_user";

DROP TYPE IF EXISTS "cauldron"."git_provider";

COMMIT;
