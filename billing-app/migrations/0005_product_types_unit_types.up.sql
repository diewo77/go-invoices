-- Create product reference tables and link products with 1:N relations
-- product_types (1) -> products (N)
-- unit_types (1) -> products (N)

-- product_types
CREATE TABLE IF NOT EXISTS product_types (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    code TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- unit_types
CREATE TABLE IF NOT EXISTS unit_types (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    symbol TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Ensure products table exists before altering (migration order might vary)
DO $$
BEGIN
    IF to_regclass('public.products') IS NULL THEN
        RETURN; -- skip if products table not present yet
    END IF;

    -- add columns if missing
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='product_type_id'
    ) THEN
        ALTER TABLE products ADD COLUMN product_type_id INTEGER;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='unit_type_id'
    ) THEN
        ALTER TABLE products ADD COLUMN unit_type_id INTEGER;
    END IF;

    -- add indexes (safe if created only once)
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE schemaname='public' AND indexname='idx_products_product_type_id'
    ) THEN
        CREATE INDEX idx_products_product_type_id ON products(product_type_id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes WHERE schemaname='public' AND indexname='idx_products_unit_type_id'
    ) THEN
        CREATE INDEX idx_products_unit_type_id ON products(unit_type_id);
    END IF;

    -- add foreign keys if not present (via constraint name guards)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
          WHERE table_name='products' AND constraint_name='fk_products_product_type_id'
    ) THEN
        ALTER TABLE products
          ADD CONSTRAINT fk_products_product_type_id
          FOREIGN KEY (product_type_id) REFERENCES product_types(id) ON UPDATE CASCADE ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
          WHERE table_name='products' AND constraint_name='fk_products_unit_type_id'
    ) THEN
        ALTER TABLE products
          ADD CONSTRAINT fk_products_unit_type_id
          FOREIGN KEY (unit_type_id) REFERENCES unit_types(id) ON UPDATE CASCADE ON DELETE SET NULL;
    END IF;
END $$ LANGUAGE plpgsql;
