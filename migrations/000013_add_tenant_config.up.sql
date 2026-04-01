CREATE TABLE IF NOT EXISTS "cauldron"."tenant_config" (
    "id" BOOLEAN NOT NULL DEFAULT TRUE PRIMARY KEY,
    "config" JSONB NOT NULL DEFAULT '{}'::jsonb,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT "single_row" CHECK ("id" = TRUE)
);

CREATE INDEX "idx_tenant_config_config" ON "cauldron"."tenant_config" USING gin (config);

INSERT INTO "cauldron"."tenant_config" ("config")
VALUES ('{}')
ON CONFLICT DO NOTHING;
