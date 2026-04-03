DROP INDEX IF EXISTS "cauldron"."idx_app_user_meta_data";
ALTER TABLE "cauldron"."app_user" DROP COLUMN IF EXISTS "meta_data";
