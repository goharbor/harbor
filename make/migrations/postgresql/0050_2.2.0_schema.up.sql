/* 
Fixes issue https://github.com/goharbor/harbor/issues/13317 
  Ensure the role_id of maintainer is 4 and the role_id of limisted guest is 5
*/
UPDATE role SET role_id=4 WHERE name='maintainer' AND role_id!=4;
UPDATE role SET role_id=5 WHERE name='limitedGuest' AND role_id!=5;

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
