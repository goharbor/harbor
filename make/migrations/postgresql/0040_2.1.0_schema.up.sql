ALTER TABLE project ADD COLUMN IF NOT EXISTS registry_id int;

CREATE TABLE IF NOT EXISTS execution (
    id SERIAL NOT NULL,
    vendor_type varchar(16) NOT NULL,
    vendor_id int,
    status varchar(16),
    status_message text,
    trigger varchar(16) NOT NULL,
    extra_attrs JSON,
    start_time timestamp DEFAULT CURRENT_TIMESTAMP,
    end_time timestamp,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS task (
    id SERIAL PRIMARY KEY NOT NULL,
    execution_id int NOT NULL,
    job_id varchar(64),
    status varchar(16) NOT NULL,
    status_code int NOT NULL,
    status_revision int,
    status_message text,
    run_count int,
    extra_attrs JSON,
    creation_time timestamp DEFAULT CURRENT_TIMESTAMP,
    start_time timestamp,
    update_time timestamp,
    end_time timestamp,
    FOREIGN KEY (execution_id) REFERENCES execution(id)
);
