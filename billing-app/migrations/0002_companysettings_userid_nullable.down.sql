-- Revert user_id to NOT NULL (will fail if null values exist)
ALTER TABLE company_settings ALTER COLUMN user_id SET NOT NULL;
