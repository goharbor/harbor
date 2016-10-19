#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sqlalchemy as sa
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from sqlalchemy.dialects import mysql

Base = declarative_base()

class User(Base):
    __tablename__ = 'user'

    user_id = sa.Column(sa.Integer, primary_key=True)
    username = sa.Column(sa.String(15), unique=True)
    email = sa.Column(sa.String(30), unique=True)
    password = sa.Column(sa.String(40), nullable=False)
    realname = sa.Column(sa.String(20), nullable=False)
    comment = sa.Column(sa.String(30))
    deleted = sa.Column(sa.Integer, nullable=False, server_default=sa.text("'0'"))
    reset_uuid = sa.Column(sa.String(40))
    salt = sa.Column(sa.String(40))
    sysadmin_flag = sa.Column(sa.Integer)
    creation_time = sa.Column(mysql.TIMESTAMP)
    update_time = sa.Column(mysql.TIMESTAMP)

class Properties(Base):
    __tablename__ = 'properties'

    k = sa.Column(sa.String(64), primary_key = True)
    v = sa.Column(sa.String(128), nullable = False)

class ProjectMember(Base):
    __tablename__ = 'project_member'

    project_id = sa.Column(sa.Integer(), primary_key = True)
    user_id = sa.Column(sa.Integer(), primary_key = True)
    role = sa.Column(sa.Integer(), nullable = False)
    creation_time = sa.Column(mysql.TIMESTAMP, nullable = True)
    update_time = sa.Column(mysql.TIMESTAMP, nullable = True)
    sa.ForeignKeyConstraint(['project_id'], [u'project.project_id'], ),
    sa.ForeignKeyConstraint(['role'], [u'role.role_id'], ),
    sa.ForeignKeyConstraint(['user_id'], [u'user.user_id'], ),

class UserProjectRole(Base):
    __tablename__ = 'user_project_role'

    upr_id = sa.Column(sa.Integer(), primary_key = True)
    user_id = sa.Column(sa.Integer(), sa.ForeignKey('user.user_id'))
    pr_id = sa.Column(sa.Integer(), sa.ForeignKey('project_role.pr_id'))
    project_role = relationship("ProjectRole")

class ProjectRole(Base):
    __tablename__ = 'project_role'

    pr_id = sa.Column(sa.Integer(), primary_key = True)
    project_id = sa.Column(sa.Integer(), nullable = False)
    role_id = sa.Column(sa.Integer(), nullable = False)
    sa.ForeignKeyConstraint(['role_id'], [u'role.role_id'])
    sa.ForeignKeyConstraint(['project_id'], [u'project.project_id'])

class Access(Base):
    __tablename__ = 'access'

    access_id = sa.Column(sa.Integer(), primary_key = True)
    access_code = sa.Column(sa.String(1))
    comment = sa.Column(sa.String(30))

class Role(Base):
    __tablename__ = 'role'

    role_id = sa.Column(sa.Integer, primary_key=True)
    role_mask = sa.Column(sa.Integer, nullable=False, server_default=sa.text("'0'"))
    role_code = sa.Column(sa.String(20))
    name = sa.Column(sa.String(20))

class Project(Base):
    __tablename__ = 'project'

    project_id = sa.Column(sa.Integer, primary_key=True)
    owner_id = sa.Column(sa.ForeignKey(u'user.user_id'), nullable=False, index=True)
    name = sa.Column(sa.String(30), nullable=False, unique=True)
    creation_time = sa.Column(mysql.TIMESTAMP)
    update_time = sa.Column(mysql.TIMESTAMP)
    deleted = sa.Column(sa.Integer, nullable=False, server_default=sa.text("'0'"))
    public = sa.Column(sa.Integer, nullable=False, server_default=sa.text("'0'"))
    owner = relationship(u'User')

class ReplicationPolicy(Base):
    __tablename__ = "replication_policy"

    id = sa.Column(sa.Integer, primary_key=True)
    name = sa.Column(sa.String(256))
    project_id = sa.Column(sa.Integer, nullable=False)
    target_id = sa.Column(sa.Integer, nullable=False)
    enabled = sa.Column(mysql.TINYINT(1), nullable=False, server_default=sa.text("'1'"))
    description = sa.Column(sa.Text)
    cron_str = sa.Column(sa.String(256))
    start_time = sa.Column(mysql.TIMESTAMP)
    creation_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP"))
    update_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"))

class ReplicationTarget(Base):
    __tablename__ = "replication_target"

    id = sa.Column(sa.Integer, primary_key=True)
    name = sa.Column(sa.String(64))
    url = sa.Column(sa.String(64))
    username = sa.Column(sa.String(40))
    password = sa.Column(sa.String(40))
    target_type = sa.Column(mysql.TINYINT(1), nullable=False, server_default=sa.text("'0'"))
    creation_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP"))
    update_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"))

class ReplicationJob(Base):
    __tablename__ = "replication_job"

    id = sa.Column(sa.Integer, primary_key=True)
    status = sa.Column(sa.String(64), nullable=False)
    policy_id = sa.Column(sa.Integer, nullable=False)
    repository = sa.Column(sa.String(256), nullable=False)
    operation = sa.Column(sa.String(64), nullable=False)
    tags = sa.Column(sa.String(16384))
    creation_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP"))
    update_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"))
    
    __table_args__ = (sa.Index('policy', "policy_id"),)

class Repository(Base):
    __tablename__ = "repository"

    repository_id = sa.Column(sa.Integer, primary_key=True)
    name = sa.Column(sa.String(255), nullable=False, unique=True)
    project_id = sa.Column(sa.Integer, nullable=False)
    owner_id = sa.Column(sa.Integer, nullable=False)
    description = sa.Column(sa.Text)
    pull_count = sa.Column(sa.Integer,server_default=sa.text("'0'"), nullable=False)
    star_count = sa.Column(sa.Integer,server_default=sa.text("'0'"), nullable=False)
    creation_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP"))
    update_time = sa.Column(mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"))


