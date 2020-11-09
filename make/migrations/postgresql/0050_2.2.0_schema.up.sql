ALTER TABLE schedule ADD COLUMN IF NOT EXISTS cron_type varchar(64);

DO $$
DECLARE
    art RECORD;
    art_size integer;
BEGIN
    FOR art IN SELECT * FROM artifact WHERE size = 0
    LOOP
      SELECT sum(size) INTO art_size FROM blob WHERE digest IN (SELECT digest_blob FROM artifact_blob WHERE digest_af=art.digest);
      UPDATE artifact SET size=art_size WHERE id = art.id;
    END LOOP;
END $$;
