-- Remove image_url column from products table
ALTER TABLE products DROP COLUMN IF EXISTS image_url;
