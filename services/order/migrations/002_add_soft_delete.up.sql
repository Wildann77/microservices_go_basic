-- Add soft delete column to orders table
ALTER TABLE orders ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index for deleted_at column
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders (deleted_at);
