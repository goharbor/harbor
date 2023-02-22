/* remove the redundant data from table artifact_blob */
delete from artifact_blob afb where not exists (select digest from blob b where b.digest = afb.digest_af);

/* replace subject_artifact_id with subject_artifact_digest*/
alter table artifact_accessory add column subject_artifact_digest varchar(1024);

DO $$
DECLARE
    acc RECORD;
    art RECORD;
BEGIN
    FOR acc IN SELECT * FROM artifact_accessory
    LOOP
        SELECT * INTO art from artifact where id = acc.subject_artifact_id;
        UPDATE artifact_accessory SET subject_artifact_digest=art.digest WHERE subject_artifact_id = art.id;
    END LOOP;
END $$;

alter table artifact_accessory drop CONSTRAINT artifact_accessory_subject_artifact_id_fkey;
alter table artifact_accessory drop CONSTRAINT unique_artifact_accessory;
alter table artifact_accessory add CONSTRAINT unique_artifact_accessory UNIQUE (artifact_id, subject_artifact_digest);
alter table artifact_accessory drop column subject_artifact_id;

