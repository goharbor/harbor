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
  
## 0.2.0

  - create table `replication_policy`
  - create table `replication_target`
  - create table `replication_job`
  - add column `repo_tag` to table `access_log`
  - alter column `repo_name` on table `access_log`
  - alter column `email` on table `user` 
