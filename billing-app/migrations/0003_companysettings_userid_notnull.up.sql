-- Ensure user_id is NOT NULL and create an initial admin user if none exists
-- 1. Create admin role if missing
INSERT INTO roles (name, description, created_at, updated_at)
SELECT 'admin','Super administrator', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name='admin');

-- 2. Create placeholder user if none exists (password should be replaced later)
INSERT INTO users (email,password,created_at,updated_at)
SELECT 'admin@example.com','changeme', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM users);

-- 3. For any company_settings rows with NULL user_id, attach the first user
UPDATE company_settings SET user_id = (
  SELECT id FROM users ORDER BY id LIMIT 1
) WHERE user_id IS NULL;

-- 4. Enforce NOT NULL
ALTER TABLE company_settings ALTER COLUMN user_id SET NOT NULL;
