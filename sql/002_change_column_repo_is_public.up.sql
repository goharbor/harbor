use registry;

ALTER TABLE repository MODIFY COLUMN is_public tinyint(2) NOT NULL DEFAULT 0;

