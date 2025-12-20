-- Revert composite unique index on products(user_id, code)
DROP INDEX IF EXISTS idx_products_user_code_unique;

-- No-op for unique index on user_id only (we don't recreate it on down).
