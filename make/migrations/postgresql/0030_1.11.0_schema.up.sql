/* TODO remove the table artifact_2 and use the artifact instead after finishing the upgrade work */
CREATE TABLE artifact_2
(
  id            SERIAL PRIMARY KEY NOT NULL,
  /* image, chart, etc */
  type          varchar(255),
  media_type    varchar(255),
  manifest_media_type varchar(255),
  project_id    int NOT NULL,
  repository_id int NOT NULL,
  digest        varchar(255) NOT NULL,
  size          bigint,  
  push_time     timestamp default CURRENT_TIMESTAMP,
  platform      varchar(255),
  extra_attrs   text,
  annotations   jsonb,
  /* when updating the data the revision MUST be checked and updated */
  revision      varchar(64) NOT NULL,
  CONSTRAINT unique_artifact_2 UNIQUE (repository_id, digest)
);

CREATE TABLE tag
(
  id            SERIAL PRIMARY KEY NOT NULL,
  repository_id int NOT NULL,
  artifact_id   int NOT NULL,
  name          varchar(255),
  push_time     timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  /* when updating the data the revision MUST be checked and updated */
  revision      varchar(64) NOT NULL,
  CONSTRAINT unique_tag UNIQUE (repository_id, name)
);

/* artifact_reference records the child artifact referenced by parent artifact */
CREATE TABLE artifact_reference
(
  id            SERIAL PRIMARY KEY NOT NULL,
  parent_id   int NOT NULL,
  child_id  int NOT NULL,
  CONSTRAINT unique_reference UNIQUE (parent_id, child_id)
);