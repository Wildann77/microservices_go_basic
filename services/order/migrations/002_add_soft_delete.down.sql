-- Remove soft delete column from orders table
DROP INDEX IF EXISTS idx_orders_deleted_at;
ALTER TABLE orders DROP COLUMN IF EXISTS deleted_at;
