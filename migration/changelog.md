# What's New in Harbor Database Schema
Changelog for harbor database schema

## 0.1.0

## 0.1.1

  - add column `creation_time` to table `user`
  - add column `sysadmin_flag` to table `user`
  - add column `update_time` to table `user`
  - add column `role_mask` to table `role`
  - drop table `user_project_role`
  - drop table `project_role`
  - delete data `sysadmin` from table `role`
  - delete data `M` from table `access`
  - add column `update_time` to table `project`
