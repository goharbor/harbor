/* TODO remove the table artifact_2 and use the artifact instead after finishing the upgrade work */
CREATE TABLE artifact_2
(
  id            SERIAL PRIMARY KEY NOT NULL,
  /* image, chart, etc */
  type          varchar(255) NOT NULL,
  media_type    varchar(255) NOT NULL,
  /* the media type of some classical image manifest can be null, so don't add the "NOT NULL" constraint*/
  manifest_media_type varchar(255),
  project_id    int NOT NULL,
  repository_id int NOT NULL,
  digest        varchar(255) NOT NULL,
  size          bigint,  
  push_time     timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  extra_attrs   text,
  annotations   jsonb,
  CONSTRAINT unique_artifact_2 UNIQUE (repository_id, digest)
);

CREATE TABLE tag
(
  id            SERIAL PRIMARY KEY NOT NULL,
  repository_id int NOT NULL,
  artifact_id   int NOT NULL,
  name          varchar(255) NOT NULL,
  push_time     timestamp default CURRENT_TIMESTAMP,
  pull_time     timestamp,
  CONSTRAINT unique_tag UNIQUE (repository_id, name)
);

/* artifact_reference records the child artifact referenced by parent artifact */
CREATE TABLE artifact_reference
(
  id          SERIAL PRIMARY KEY NOT NULL,
  parent_id   int NOT NULL,
  child_id    int NOT NULL,
  platform    varchar(255),
  CONSTRAINT  unique_reference UNIQUE (parent_id, child_id)
);