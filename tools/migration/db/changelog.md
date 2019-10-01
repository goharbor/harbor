# What's New in Harbor Database Schema
Changelog for harbor database schema

## 0.1.0

## 0.1.1

  - create table `project_member`
  - create table `schema_version`
  - drop table `user_project_role`
  - drop table `project_role`
  - add column `creation_time` to table `user`
  - add column `sysadmin_flag` to table `user`
  - add column `update_time` to table `user`
  - add column `role_mask` to table `role`
  - add column `update_time` to table `project`
  - delete data `AMDRWS` from table `role`
  - delete data `A` from table `access`

## 0.3.0

  - create table `replication_policy`
  - create table `replication_target`
  - create table `replication_job`
  - add column `repo_tag` to table `access_log`
  - alter column `repo_name` on table `access_log`
  - alter column `email` on table `user`

## 0.4.0

  - add index `pid_optime (project_id, op_time)` on table `access_log`
  - add index `poid_uptime (policy_id, update_time)` on table `replication_job`
  - add column `deleted` to table `replication_policy`
  - alter column `username` on table `user`: varchar(15)->varchar(32)
  - alter column `password` on table `replication_target`: varchar(40)->varchar(128)
  - alter column `email` on table `user`: varchar(128)->varchar(255)
  - alter column `name` on table `project`: varchar(30)->varchar(41)
  - create table `repository`
  - alter column `password` on table `replication_target`: varchar(40)->varchar(128)

## 1.2.0

  - delete column `owner_id` from table `repository`
  - delete column `user_id` from table `access_log`
  - delete foreign key (user_id) references user(user_id)from table `access_log`
  - delete foreign key (project_id) references project(project_id)from table `access_log`
  - add column `username` varchar (32) to table `access_log`
  - alter column `realname` on table `user`: varchar(20)->varchar(255)
  - create table `img_scan_job`
  - create table `img_scan_overview`
  - create table `clair_vuln_timestamp`

## 1.3.0

  - create table `project_metadata`
  - insert data into table `project_metadata`
  - delete column `public` from table `project`
  - add column `insecure` to table `replication_target`

## 1.4.0

  - add column `filters` to table `replication_policy`
  - add column `replicate_deletion` to table `replication_policy`
  - create table `replication_immediate_trigger`
  - add pk `id` to table `properties`
  - remove pk index from column 'k' of table `properties`
  - alter `name` length from 41 to 256 of table `project`

## 1.5.0

  - create table `harbor_label`
  - create table `harbor_resource_label`
  - create table `user_group`
  - modify table `project_member` use `id` as PK and add column `entity_type` to indicate if the member is user or group.
  - add `job_uuid` column to `replication_job` and `img_scan_job`
  - add index `poid_status` in table replication_job
  - add index `idx_status`, `idx_status`, `idx_digest`, `idx_repository_tag` in table img_scan_job

## 1.6.0

  - add `deleted` column to table `harbor_label`

## 1.7.0

  - alter column `v` on table `properties`: varchar(128)->varchar(1024)
