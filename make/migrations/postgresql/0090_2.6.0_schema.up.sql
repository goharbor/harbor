/* Correct project_metadata.public value, should only be true or false, other invaild value will be rewrite to false */
UPDATE project_metadata SET value='false' WHERE name='public' AND value NOT IN('true', 'false');

/*
System Artifact Manager
Github proposal link : https://github.com/goharbor/community/pull/181
*/
 CREATE TABLE IF NOT EXISTS system_artifact (
        id SERIAL NOT NULL PRIMARY KEY,
        repository varchar(256) NOT NULL,
        digest varchar(255) NOT NULL DEFAULT '' ,
        size bigint NOT NULL DEFAULT 0 ,
        vendor varchar(255) NOT NULL DEFAULT '' ,
        type varchar(255) NOT NULL DEFAULT '' ,
        create_time timestamp default CURRENT_TIMESTAMP,
        extra_attrs text NOT NULL DEFAULT '' ,
        UNIQUE ("repository", "digest",  "vendor")
);