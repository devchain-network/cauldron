CREATE INDEX "idx_github_delivery_id" ON
    "cauldron"."github" (delivery_id);

CREATE INDEX "idx_github_payload_check_run_conclusion" ON
    "cauldron"."github" USING btree ((payload->'check_run'->>'conclusion'))
    WHERE event = 'check_run';

CREATE INDEX "idx_github_payload_check_suite_conclusion" ON
    "cauldron"."github" USING btree ((payload->'check_suite'->>'conclusion'))
    WHERE event = 'check_suite';
