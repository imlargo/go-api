-- Initialize the database with basic setup
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create indexes for better performance (these should match your GORM models)
-- This is just an example - adjust based on your actual schema

-- Example: Create an index on user email for faster lookups
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Example: Create an index on created_at for time-based queries  
-- CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Add any other initialization SQL here