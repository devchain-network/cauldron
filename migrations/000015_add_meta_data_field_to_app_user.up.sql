ALTER TABLE "cauldron"."app_user"
    ADD COLUMN "meta_data" JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX "idx_app_user_meta_data" ON "cauldron"."app_user" USING gin ("meta_data");
