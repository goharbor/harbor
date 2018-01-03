CREATE TABLE "private_keys" (
  "id" serial PRIMARY KEY,
  "created_at" timestamp NULL DEFAULT NULL,
  "updated_at" timestamp NULL DEFAULT NULL,
  "deleted_at" timestamp NULL DEFAULT NULL,
  "key_id" varchar(255) NOT NULL,
  "encryption_alg" varchar(255) NOT NULL,
  "keywrap_alg" varchar(255) NOT NULL,
  "algorithm" varchar(50) NOT NULL,
  "passphrase_alias" varchar(50) NOT NULL,
  "public" bytea NOT NULL,
  "private" bytea NOT NULL,
  "gun" varchar(255) NOT NULL,
  "role" varchar(255) NOT NULL,
  "last_used" timestamp NULL DEFAULT NULL,
  CONSTRAINT "key_id" UNIQUE ("key_id"),
  CONSTRAINT "key_id_2" UNIQUE ("key_id","algorithm")
);
