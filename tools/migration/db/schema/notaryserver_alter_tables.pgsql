\c notaryserver;

ALTER TABLE tuf_files OWNER TO server;
ALTER SEQUENCE tuf_files_id_seq OWNER TO server;
ALTER TABLE change_category OWNER TO server;
ALTER TABLE changefeed OWNER TO server;
ALTER SEQUENCE changefeed_id_seq OWNER TO server;
ALTER TABLE schema_migrations OWNER TO server;