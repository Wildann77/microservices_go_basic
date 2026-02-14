-- Add soft delete column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index for deleted_at column
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
