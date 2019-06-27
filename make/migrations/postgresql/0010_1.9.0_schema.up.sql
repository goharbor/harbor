/* add table for CVE whitelist */
CREATE TABLE cve_whitelist (
    id SERIAL PRIMARY KEY NOT NULL,
    project_id int,
    creation_time timestamp default CURRENT_TIMESTAMP,
    update_time timestamp default CURRENT_TIMESTAMP,
    expires_at bigint,
    items text NOT NULL,
    UNIQUE (project_id)
);