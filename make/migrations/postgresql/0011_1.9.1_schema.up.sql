ALTER TABLE harbor_user ADD COLUMN password_version varchar(16) Default 'sha256';
UPDATE harbor_user SET password_version = 'sha1';