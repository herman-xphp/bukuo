-- Drop indexes
DROP INDEX IF EXISTS idx_products_active;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_type;
DROP INDEX IF EXISTS idx_products_company;

DROP INDEX IF EXISTS idx_product_categories_parent;
DROP INDEX IF EXISTS idx_product_categories_company;

DROP INDEX IF EXISTS idx_units_company;

-- Drop tables
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS units;
