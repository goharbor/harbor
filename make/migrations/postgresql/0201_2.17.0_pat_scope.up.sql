-- Add scope column to personal_access_token table
ALTER TABLE personal_access_token ADD COLUMN IF NOT EXISTS scope TEXT;