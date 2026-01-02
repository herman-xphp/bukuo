DROP INDEX IF EXISTS idx_product_stocks_product;
DROP INDEX IF EXISTS idx_product_stocks_company;
DROP TABLE IF EXISTS product_stocks;

DROP INDEX IF EXISTS idx_inv_tx_date;
DROP INDEX IF EXISTS idx_inv_tx_warehouse;
DROP INDEX IF EXISTS idx_inv_tx_product;
DROP INDEX IF EXISTS idx_inv_tx_company;
DROP TABLE IF EXISTS inventory_transactions;

DROP INDEX IF EXISTS idx_warehouses_company;
DROP TABLE IF EXISTS warehouses;
