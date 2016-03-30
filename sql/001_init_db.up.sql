use registry;

CREATE TABLE repository (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL DEFAULT '',
  project_name varchar(255) NOT NULL DEFAULT '',
  project_id bigint(11) NOT NULL,
  created_at datetime DEFAULT NULL,
  updated_at datetime DEFAULT NULL,
  user_name varchar(155) NOT NULL,
  category varchar(255),
  is_public tinyint(2) NOT NULL DEFAULT 1,
  latest_tag varchar(255) NOT NULL DEFAULT 'latest',
  description varchar(512),

  PRIMARY KEY(`id`),
  KEY `index_repository_project_id` (`project_id`)
);

CREATE TABLE tag (
  `id` bigint(11) NOT NULL AUTO_INCREMENT,
  project_id bigint(11) NOT NULL,
  repository_id bigint(11) NOT NULL,
  version varchar(255) NOT NULL DEFAULT '',
  created_at datetime DEFAULT NULL,
  updated_at datetime DEFAULT NULL,
  PRIMARY KEY(`id`),
  KEY `index_tag_reposotory_id` (`repository_id`)
);
