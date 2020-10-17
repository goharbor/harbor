/*
Fixes https://github.com/goharbor/harbor/issues/12827
 After user migrates Harbor from v2.0.2, user got 404 when to pull specific images, and no work after push the same images again.

 Fix:
 1, If the issue is caused by missing repository data, this fix can revert the missing repository data and all things should be fine.
 2, If the issue is caused by missing blob data, this fix can revert the missing repository data and still left the media type of artifact
    as 'UNKNOWN', which leads the meta data and build history of the image cannot be shown in UI. User can delete and push the image again to
    resolve it.
*/

/* Delete the duplicate tags if user re-tag & re-push the missing images */
DELETE FROM tag
WHERE id NOT IN
    (SELECT MAX(id) AS id
    FROM (SELECT t.*, art.repository_name FROM artifact AS art JOIN tag AS t ON art.id = t.artifact_id) t1
    GROUP BY t1.name, t1.repository_name);

/* Insert the missing repository records */
INSERT INTO repository (name, project_id)
    SELECT DISTINCT repository_name, project_id FROM artifact WHERE repository_id<0 AND
    repository_name NOT IN (SELECT name from repository);

/* Update the repository id of artifact records */
UPDATE artifact AS art
    SET repository_id=repo.repository_id
    FROM repository AS repo
    WHERE art.repository_name=repo.name AND art.repository_id!=repo.repository_id;

/* Update the media type of artifact records */
UPDATE artifact AS art
    SET manifest_media_type=blob.content_type,
    media_type=(
    CASE
        /*v2 manifest*/
        WHEN blob.content_type='application/vnd.docker.distribution.manifest.v2+json' THEN
            'application/vnd.docker.container.image.v1+json'
        /*manifest list*/
        WHEN blob.content_type='application/vnd.docker.distribution.manifest.list.v2+json' THEN
            'application/vnd.docker.distribution.manifest.list.v2+json'
        /*v1 manifest*/
        ELSE
            'application/vnd.docker.distribution.manifest.v1+prettyjws'
    END
    )
    FROM blob AS blob
    WHERE art.media_type='UNKNOWN' AND art.digest=blob.digest;

/* update tag records with negative repository id */
UPDATE tag SET
   repository_id=art.repository_id
   FROM artifact as art
   WHERE tag.artifact_id=art.id AND tag.repository_id!=art.repository_id;


