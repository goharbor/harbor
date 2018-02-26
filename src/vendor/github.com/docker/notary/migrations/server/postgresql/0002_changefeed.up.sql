CREATE TABLE "change_category" (
    "category" VARCHAR(20) PRIMARY KEY
);

INSERT INTO "change_category" VALUES ('update'), ('deletion');

CREATE TABLE "changefeed" (
    "id" serial PRIMARY KEY,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP,
    "gun" varchar(255) NOT NULL,
    "version" integer NOT NULL,
    "sha256" CHAR(64) DEFAULT NULL,
    "category" VARCHAR(20) NOT NULL DEFAULT 'update' REFERENCES "change_category"
);

CREATE INDEX "idx_changefeed_gun" ON "changefeed" ("gun");

INSERT INTO "changefeed" (
        "created_at",
        "gun",
        "version",
        "sha256" 
    ) (SELECT
        "created_at",
        "gun",
        "version",
        "sha256"
    FROM
        "tuf_files"
    WHERE
        "role" = 'timestamp'
    ORDER BY
        "created_at" ASC
);
