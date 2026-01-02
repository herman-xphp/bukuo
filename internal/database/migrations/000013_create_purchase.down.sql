DROP INDEX IF EXISTS idx_grn_company;
DROP INDEX IF EXISTS idx_purchase_invoices_supplier;
DROP INDEX IF EXISTS idx_purchase_invoices_company;
DROP INDEX IF EXISTS idx_purchase_orders_supplier;
DROP INDEX IF EXISTS idx_purchase_orders_company;

DROP TABLE IF EXISTS purchase_return_lines;
DROP TABLE IF EXISTS purchase_returns;
DROP TABLE IF EXISTS purchase_invoice_lines;
DROP TABLE IF EXISTS purchase_invoices;
DROP TABLE IF EXISTS grn_lines;
DROP TABLE IF EXISTS goods_received_notes;
DROP TABLE IF EXISTS purchase_order_lines;
DROP TABLE IF EXISTS purchase_orders;
DROP TABLE IF EXISTS purchase_request_lines;
DROP TABLE IF EXISTS purchase_requests;
