 --force an upgrade of the schema-- REMOVE once dev complete
 DROP TABLE IF EXISTS "scan_report_v2";
 DROP TABLE IF EXISTS "vulnerability_record_v2";
 DROP TABLE IF EXISTS "report_vulnerability_record_v2";
 -----------------------------
 
 -- --------------------------------------------------
    --  Table Structure for `main.Report`
    -- --------------------------------------------------
    CREATE TABLE IF NOT EXISTS "scan_report_v2" (
        "id" serial NOT NULL PRIMARY KEY,
        "uuid" text NOT NULL DEFAULT ''  UNIQUE,
        "digest" text NOT NULL DEFAULT '' ,
        "registration_uuid" text NOT NULL DEFAULT '' ,
        "mime_type" text NOT NULL DEFAULT '' ,
        "job_id" text NOT NULL DEFAULT '' ,
        "track_id" text NOT NULL DEFAULT '' ,
        "requester" text NOT NULL DEFAULT '' ,
        "status" text NOT NULL DEFAULT '' ,
        "status_code" integer NOT NULL DEFAULT 0 ,
        "status_rev" bigint NOT NULL DEFAULT 0 ,
        "start_time" timestamp DEFAULT CURRENT_TIMESTAMP,
        "end_time" timestamp DEFAULT CURRENT_TIMESTAMP,
        UNIQUE ("uuid"),
        UNIQUE ("digest", "registration_uuid", "mime_type")
    );

    -- --------------------------------------------------
    --  Table Structure for `main.VulnerabilityRecord`
    -- --------------------------------------------------
    CREATE TABLE IF NOT EXISTS "vulnerability_record_v2" (
        "id" serial NOT NULL PRIMARY KEY,
        "cve_id" text NOT NULL DEFAULT '' ,
        "registration_uuid" text NOT NULL DEFAULT '',
        "digest" text NOT NULL DEFAULT '' ,
        "report_uuid" text NOT NULL DEFAULT '' ,
        "package" text NOT NULL DEFAULT '' ,
        "package_version" text NOT NULL DEFAULT '' ,
        "package_type" text NOT NULL DEFAULT '' ,
        "severity" text NOT NULL DEFAULT '' ,
        "fixed_version" text,
        "urls" text,
        "cve3_score" text,
        "cve2_score" text,
        "cvss3_vector" text,
        "cvss2_vector" text,
        "description" text,
        "vendorattributes" json,
        UNIQUE ("cve_id", "registration_uuid")
    );

    -- --------------------------------------------------
    --  Table Structure for `main.ReportVulnerabilityRecord`
    -- --------------------------------------------------
    CREATE TABLE IF NOT EXISTS "report_vulnerability_record_v2" (
        "id" serial NOT NULL PRIMARY KEY,
        "report_uuid" text NOT NULL DEFAULT '' ,
        "vuln_record_id" bigint NOT NULL DEFAULT 0 ,
        UNIQUE ("report_uuid", "vuln_record_id")
    );
