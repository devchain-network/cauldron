CREATE INDEX IF NOT EXISTS "idx_github_user_login" ON "cauldron"."github" USING btree (user_login);
