-- Revert NOT NULL on user_id (makes it nullable again)
ALTER TABLE company_settings ALTER COLUMN user_id DROP NOT NULL;
