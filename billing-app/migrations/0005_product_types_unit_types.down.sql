-- Drop relations and reference tables (reverse of 0005)
DO $$
BEGIN
    IF to_regclass('public.products') IS NOT NULL THEN
        -- drop FKs if exist
        IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name='products' AND constraint_name='fk_products_product_type_id') THEN
            ALTER TABLE products DROP CONSTRAINT fk_products_product_type_id;
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name='products' AND constraint_name='fk_products_unit_type_id') THEN
            ALTER TABLE products DROP CONSTRAINT fk_products_unit_type_id;
        END IF;
        -- drop indexes
        DROP INDEX IF EXISTS idx_products_product_type_id;
        DROP INDEX IF EXISTS idx_products_unit_type_id;
        -- drop columns
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='product_type_id') THEN
            ALTER TABLE products DROP COLUMN product_type_id;
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='unit_type_id') THEN
            ALTER TABLE products DROP COLUMN unit_type_id;
        END IF;
    END IF;
END $$ LANGUAGE plpgsql;

DROP TABLE IF EXISTS product_types;
DROP TABLE IF EXISTS unit_types;
