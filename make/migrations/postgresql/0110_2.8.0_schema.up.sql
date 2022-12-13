/* remove the redundant data from table artifact_blob */
delete from artifact_blob afb where not exists (select digest from blob b where b.digest = afb.digest_af);