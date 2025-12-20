-- Make company_settings.user_id nullable (already nullable logically) and keep FK
ALTER TABLE company_settings ALTER COLUMN user_id DROP NOT NULL;
