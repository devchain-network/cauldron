ALTER TABLE "cauldron"."github_org_settings"
    ADD COLUMN "meta_data" JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX "idx_github_org_settings_meta_data" ON "cauldron"."github_org_settings" USING gin ("meta_data");
