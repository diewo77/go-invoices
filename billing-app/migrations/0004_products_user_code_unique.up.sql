-- Ensure composite unique index on (user_id, code) for products
-- and drop any accidental unique index on user_id alone.

-- Only proceed if products table exists
DO $$
BEGIN
    IF to_regclass('public.products') IS NULL THEN
        RETURN; -- products table not created yet; skip safely
    END IF;

    -- Drop potential wrong unique index if it exists
    IF EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE schemaname = 'public' AND indexname = 'idx_products_user_id_unique'
    ) THEN
        EXECUTE 'DROP INDEX IF EXISTS idx_products_user_id_unique';
    END IF;

    -- Also drop GORM automatic unique indices that might have different names
    -- Attempt to drop by pattern; harmless if not present
    PERFORM 1;
    FOR r IN SELECT indexname FROM pg_indexes WHERE schemaname='public' AND indexname LIKE 'products_user_id_%_key' LOOP
        EXECUTE format('DROP INDEX IF EXISTS %I', r.indexname);
    END LOOP;

    -- Create composite unique if missing
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE schemaname='public' AND indexname='idx_products_user_code_unique'
    ) THEN
        EXECUTE 'CREATE UNIQUE INDEX idx_products_user_code_unique ON products (user_id, code)';
    END IF;
END $$ LANGUAGE plpgsql;
