CREATE INDEX "idx_github_payload_sender_login" ON 
    "cauldron"."github" USING btree ((payload->'sender'->>'login'));

CREATE INDEX "idx_github_payload_pr_user_login" ON 
    "cauldron"."github" USING btree (((payload->'pull_request'->'user'->>'login')))
    WHERE event = 'pull_request';
