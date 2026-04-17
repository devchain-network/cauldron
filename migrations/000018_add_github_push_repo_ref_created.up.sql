CREATE INDEX IF NOT EXISTS "idx_github_push_repo_ref_created" ON
    "cauldron"."github" (((payload->'repository'->>'name')), ((payload->>'ref')), created_at) 
    WHERE event='push';
