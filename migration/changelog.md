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
