/*Table for keeping the plug scanner registration*/
CREATE TABLE scanner_registration
(
    id SERIAL PRIMARY KEY NOT NULL,
    uuid VARCHAR(64) UNIQUE NOT NULL,
    url VARCHAR(256) UNIQUE NOT NULL,
    name VARCHAR(128) UNIQUE NOT NULL,
    description VARCHAR(1024) NULL,
    auth VARCHAR(16) NOT NULL,
    access_cred VARCHAR(512) NULL,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    use_internal_addr BOOLEAN NOT NULL DEFAULT FALSE,
    immutable BOOLEAN NOT NULL DEFAULT FALSE,
    skip_cert_verify BOOLEAN NOT NULL DEFAULT FALSE,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

/*Table for keeping the scan report. The report details are stored as JSON*/
CREATE TABLE scan_report
(
    id SERIAL PRIMARY KEY NOT NULL,
    uuid VARCHAR(64) UNIQUE NOT NULL,
    digest VARCHAR(256) NOT NULL,
    registration_uuid VARCHAR(64) NOT NULL,
    mime_type VARCHAR(256) NOT NULL,
    job_id VARCHAR(64),
    track_id VARCHAR(64),
    request_id VARCHAR(64),
    status VARCHAR(1024) NOT NULL,
    status_code INTEGER DEFAULT 0,
    status_rev BIGINT DEFAULT 0,
    report JSON,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(digest, registration_uuid, mime_type)
);

/** Add table for immutable tag  **/
CREATE TABLE immutable_tag_rule
(
  id SERIAL PRIMARY KEY NOT NULL,
  project_id int NOT NULL,
  tag_filter text,
  disabled BOOLEAN NOT NULL DEFAULT FALSE,
  creation_time timestamp default CURRENT_TIMESTAMP
);

ALTER TABLE robot ADD COLUMN visible boolean DEFAULT true NOT NULL;

/** Drop the unused vul related tables **/
DROP INDEX IF EXISTS idx_status;
DROP INDEX IF EXISTS idx_digest;
DROP INDEX IF EXISTS idx_uuid;
DROP INDEX IF EXISTS idx_repository_tag;
DROP TRIGGER IF EXISTS img_scan_job_update_time_at_modtime ON img_scan_job;
DROP TABLE IF EXISTS img_scan_job;

DROP TRIGGER IF EXISTS TRIGGER ON img_scan_overview;
DROP TABLE IF EXISTS img_scan_overview;

DROP TABLE IF EXISTS clair_vuln_timestamp;

/* Add limited guest role */
INSERT INTO role (role_code, name) VALUES ('LRS', 'limitedGuest');
