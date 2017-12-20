CREATE TABLE "tuf_files" (
  "id" serial PRIMARY KEY,
  "created_at" timestamp NULL DEFAULT NULL,
  "updated_at" timestamp NULL DEFAULT NULL,
  "deleted_at" timestamp NULL DEFAULT NULL,
  "gun" varchar(255) NOT NULL,
  "role" varchar(255) NOT NULL,
  "version" integer NOT NULL,
  "data" bytea NOT NULL,
  "sha256" char(64) DEFAULT NULL,
  UNIQUE ("gun","role","version")
);

CREATE INDEX tuf_files_sha256_idx ON tuf_files(sha256);
