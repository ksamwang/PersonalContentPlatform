ALTER TABLE webhook_endpoints DROP CONSTRAINT IF EXISTS webhook_endpoints_secret_ref_check;
ALTER TABLE webhook_endpoints ADD CONSTRAINT webhook_endpoints_secret_check CHECK (secret_ref <> '' OR secret_value <> '');
