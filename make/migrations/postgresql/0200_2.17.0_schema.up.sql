-- Personal Access Tokens table for all auth modes (OIDC, DB, LDAP)
CREATE TABLE IF NOT EXISTS personal_access_token (
    id            BIGSERIAL PRIMARY KEY NOT NULL,
    user_id       integer NOT NULL REFERENCES harbor_user(user_id),
    name          varchar(255) NOT NULL,
    secret        varchar(2048) NOT NULL,
    salt          varchar(64) NOT NULL,
    description   varchar(1024),
    expires_at    bigint NOT NULL DEFAULT -1,
    last_used_at  bigint,
    disabled      boolean NOT NULL DEFAULT false,
    is_legacy     boolean NOT NULL DEFAULT false,
    creation_time timestamp DEFAULT CURRENT_TIMESTAMP,
    update_time   timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_pat_name UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_pat_user_id ON personal_access_token (user_id);
CREATE INDEX IF NOT EXISTS idx_pat_disabled ON personal_access_token (disabled);

CREATE TRIGGER pat_update_time_at_modtime
    BEFORE UPDATE ON personal_access_token FOR EACH ROW
    EXECUTE PROCEDURE update_update_time_at_column();
