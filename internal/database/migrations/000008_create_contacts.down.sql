-- Drop indexes
DROP INDEX IF EXISTS idx_contacts_active;
DROP INDEX IF EXISTS idx_contacts_name;
DROP INDEX IF EXISTS idx_contacts_type;
DROP INDEX IF EXISTS idx_contacts_company;

-- Drop table
DROP TABLE IF EXISTS contacts;
