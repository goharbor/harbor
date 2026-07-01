-- Remove scope column from personal_access_token table
ALTER TABLE personal_access_token DROP COLUMN IF EXISTS scope;
