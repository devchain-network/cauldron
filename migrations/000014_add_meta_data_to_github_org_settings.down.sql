DROP INDEX IF EXISTS "cauldron"."idx_github_org_settings_meta_data";

ALTER TABLE "cauldron"."github_org_settings"
    DROP COLUMN IF EXISTS "meta_data";
