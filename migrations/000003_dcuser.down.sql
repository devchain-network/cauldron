BEGIN;

DROP INDEX IF EXISTS idx_github_user_meta_data;
DROP INDEX IF EXISTS idx_github_user_user_id;
DROP INDEX IF EXISTS idx_github_user_user_login;

DROP TABLE IF EXISTS github_user;
DROP TABLE IF EXISTS dcuser;

DROP TYPE IF EXISTS git_provider;

COMMIT;
