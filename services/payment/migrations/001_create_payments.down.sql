-- Drop trigger
DROP TRIGGER IF EXISTS update_payments_updated_at_trigger ON payments;
DROP FUNCTION IF EXISTS update_payments_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_payments_transaction_id;
DROP INDEX IF EXISTS idx_payments_user_status;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_order_id;

-- Drop table
DROP TABLE IF EXISTS payments;