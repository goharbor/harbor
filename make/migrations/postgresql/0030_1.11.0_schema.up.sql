/* TODO remove the table artifact_2 and use the artifact instead after finishing the upgrade work */
CREATE TABLE artifact_2
(
  id            SERIAL PRIMARY KEY NOT NULL,
  /* image, chart, etc */
  type          varchar(255) NOT NULL,
  media_type    varchar(255) NOT NULL,
  manifest_media_type varchar(255) NOT NULL,
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
  /* TODO replace artifact_2 after finishing the upgrade work */
  FOREIGN KEY (artifact_id) REFERENCES artifact_2(id),
  CONSTRAINT unique_tag UNIQUE (repository_id, name)
);

/* artifact_reference records the child artifact referenced by parent artifact */
CREATE TABLE artifact_reference
(
  id          SERIAL PRIMARY KEY NOT NULL,
  parent_id   int NOT NULL,
  child_id    int NOT NULL,
  platform    varchar(255),
  /* TODO replace artifact_2 after finishing the upgrade work */
  FOREIGN KEY (parent_id) REFERENCES artifact_2(id),
  FOREIGN KEY (child_id) REFERENCES artifact_2(id),
  CONSTRAINT  unique_reference UNIQUE (parent_id, child_id)
);


/* TODO upgrade: how about keep the table "harbor_resource_label" only for helm v2 chart and use the new table for artifact label reference? */
/* label_reference records the labels added to the artifact */
CREATE TABLE label_reference (
 id SERIAL PRIMARY KEY NOT NULL,
 label_id int NOT NULL,
 artifact_id int NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 FOREIGN KEY (label_id) REFERENCES harbor_label(id),
 /* TODO replace artifact_2 after finishing the upgrade work */
 FOREIGN KEY (artifact_id) REFERENCES artifact_2(id),
 CONSTRAINT unique_label_reference UNIQUE (label_id,artifact_id)
 );
