#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sqlalchemy as sa
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from sqlalchemy.dialects import postgresql
from sqlalchemy.sql import func
import datetime

Base = declarative_base()


class User(Base):
    __tablename__ = 'harbor_user'

    user_id = sa.Column(sa.Integer, primary_key=True)
    username = sa.Column(sa.String(255), unique=True)
    email = sa.Column(sa.String(255), unique=True)
    password = sa.Column(sa.String(40), nullable=False)
    realname = sa.Column(sa.String(255), nullable=False)
    comment = sa.Column(sa.String(30))
    deleted = sa.Column(sa.Boolean, nullable=False, server_default='false')
    reset_uuid = sa.Column(sa.String(40))
    salt = sa.Column(sa.String(40))
    sysadmin_flag = sa.Column(sa.Boolean, nullable=False, server_default='false')
    creation_time = sa.Column(sa.TIMESTAMP)
    update_time = sa.Column(sa.TIMESTAMP)


class UserGroup(Base):
    __tablename__ = 'user_group'

    id = sa.Column(sa.Integer, primary_key=True)
    group_name = sa.Column(sa.String(255), nullable = False)
    group_type = sa.Column(sa.SmallInteger, server_default=sa.text("'0'"))
    ldap_group_dn = sa.Column(sa.String(512), nullable=False)
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class Properties(Base):
    __tablename__ = 'properties'

    id = sa.Column(sa.Integer, primary_key=True)
    k = sa.Column(sa.String(64), unique=True)
    v = sa.Column(sa.String(128), nullable = False)


class ProjectMember(Base):
    __tablename__ = 'project_member'

    id = sa.Column(sa.Integer, primary_key=True)
    project_id = sa.Column(sa.Integer(), nullable=False)
    entity_id = sa.Column(sa.Integer(), nullable=False)
    entity_type = sa.Column(sa.String(1), nullable=False)
    role = sa.Column(sa.Integer(), nullable = False)
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))

    __table_args__ = (sa.UniqueConstraint('project_id', 'entity_id', 'entity_type', name='unique_name_and_scope'),)


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
    owner_id = sa.Column(sa.ForeignKey(u'harbor_user.user_id'), nullable=False, index=True)
    name = sa.Column(sa.String(255), nullable=False, unique=True)
    creation_time = sa.Column(sa.TIMESTAMP)
    update_time = sa.Column(sa.TIMESTAMP)
    deleted = sa.Column(sa.Boolean, nullable=False, server_default='false')
    owner = relationship(u'User')


class ProjectMetadata(Base):
    __tablename__ = 'project_metadata'

    id = sa.Column(sa.Integer, primary_key=True)
    project_id = sa.Column(sa.ForeignKey(u'project.project_id'), nullable=False)
    name = sa.Column(sa.String(255), nullable=False)
    value = sa.Column(sa.String(255))
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    deleted = sa.Column(sa.Boolean, nullable=False, server_default='false')

    __table_args__ = (sa.UniqueConstraint('project_id', 'name', name='unique_project_id_and_name'),)


class ReplicationPolicy(Base):
    __tablename__ = "replication_policy"

    id = sa.Column(sa.Integer, primary_key=True)
    name = sa.Column(sa.String(256))
    project_id = sa.Column(sa.Integer, nullable=False)
    target_id = sa.Column(sa.Integer, nullable=False)
    enabled = sa.Column(sa.Boolean, nullable=False, server_default='true')
    description = sa.Column(sa.Text)
    cron_str = sa.Column(sa.String(256))
    filters = sa.Column(sa.String(1024))
    replicate_deletion = sa.Column(sa.Boolean, nullable=False, server_default='false')
    start_time = sa.Column(sa.TIMESTAMP)
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class ReplicationTarget(Base):
    __tablename__ = "replication_target"

    id = sa.Column(sa.Integer, primary_key=True)
    name = sa.Column(sa.String(64))
    url = sa.Column(sa.String(64))
    username = sa.Column(sa.String(255))
    password = sa.Column(sa.String(128))
    target_type = sa.Column(sa.SmallInteger, nullable=False, server_default=sa.text("'0'"))
    insecure = sa.Column(sa.Boolean, nullable=False, server_default='false')
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class ReplicationJob(Base):
    __tablename__ = "replication_job"

    id = sa.Column(sa.Integer, primary_key=True)
    status = sa.Column(sa.String(64), nullable=False)
    policy_id = sa.Column(sa.Integer, nullable=False)
    repository = sa.Column(sa.String(256), nullable=False)
    operation = sa.Column(sa.String(64), nullable=False)
    tags = sa.Column(sa.String(16384))
    job_uuid = sa.Column(sa.String(64))
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))

    __table_args__ = (sa.Index('policy', 'policy_id'),)


class ReplicationImmediateTrigger(Base):
    __tablename__ = 'replication_immediate_trigger'

    id = sa.Column(sa.Integer, primary_key=True)
    policy_id = sa.Column(sa.Integer, nullable=False)
    namespace = sa.Column(sa.String(256), nullable=False)
    on_push = sa.Column(sa.Boolean, nullable=False, server_default='false')
    on_deletion = sa.Column(sa.Boolean, nullable=False, server_default='false')
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class Repository(Base):
    __tablename__ = "repository"

    repository_id = sa.Column(sa.Integer, primary_key=True)
    name = sa.Column(sa.String(255), nullable=False, unique=True)
    project_id = sa.Column(sa.Integer, nullable=False)
    owner_id = sa.Column(sa.Integer, nullable=False)
    description = sa.Column(sa.Text)
    pull_count = sa.Column(sa.Integer,server_default=sa.text("'0'"), nullable=False)
    star_count = sa.Column(sa.Integer,server_default=sa.text("'0'"), nullable=False)
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class AccessLog(Base):
    __tablename__ = "access_log"

    log_id = sa.Column(sa.Integer, primary_key=True)
    username = sa.Column(sa.String(255), nullable=False)
    project_id = sa.Column(sa.Integer, nullable=False)
    repo_name = sa.Column(sa.String(256))
    repo_tag = sa.Column(sa.String(128))
    GUID = sa.Column(sa.String(64))
    operation = sa.Column(sa.String(20))
    op_time = sa.Column(sa.TIMESTAMP)

    __table_args__ = (sa.Index('project_id', "op_time"),)


class ImageScanJob(Base):
    __tablename__ = "img_scan_job"

    id = sa.Column(sa.Integer, nullable=False, primary_key=True)
    status = sa.Column(sa.String(64), nullable=False)
    repository = sa.Column(sa.String(256), nullable=False)
    tag = sa.Column(sa.String(128), nullable=False)
    digest = sa.Column(sa.String(128))
    job_uuid = sa.Column(sa.String(64))
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class ImageScanOverview(Base):
    __tablename__ = "img_scan_overview"

    id = sa.Column(sa.Integer, nullable=False, primary_key=True)
    image_digest = sa.Column(sa.String(128), nullable=False)
    scan_job_id = sa.Column(sa.Integer, nullable=False)
    severity = sa.Column(sa.Integer, nullable=False, server_default=sa.text("'0'"))
    components_overview = sa.Column(sa.String(2048))
    details_key = sa.Column(sa.String(128))
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


class ClairVulnTimestamp(Base):
    __tablename__ = "clair_vuln_timestamp"

    id = sa.Column(sa.Integer, nullable=False, primary_key=True)
    namespace = sa.Column(sa.String(128), nullable=False, unique=True)
    last_update = sa.Column(sa.TIMESTAMP)


class HarborLabel(Base):
    __tablename__ = "harbor_label"

    id = sa.Column(sa.Integer, nullable=False, primary_key=True)
    name = sa.Column(sa.String(128), nullable=False)
    description = sa.Column(sa.Text)
    color = sa.Column(sa.String(16))
    level = sa.Column(sa.String(1), nullable=False)
    scope = sa.Column(sa.String(1), nullable=False)
    project_id = sa.Column(sa.Integer, nullable=False)
    deleted = sa.Column(sa.Boolean, nullable=False, server_default='false')
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))

    __table_args__ = (sa.UniqueConstraint('name', 'scope', 'project_id', name='unique_label'),)


class HarborResourceLabel(Base):
    __tablename__ = 'harbor_resource_label'

    id = sa.Column(sa.Integer, primary_key=True)
    label_id = sa.Column(sa.Integer, nullable=False)
    resource_id =  sa.Column(sa.Integer)
    resource_name = sa.Column(sa.String(256))
    resource_type = sa.Column(sa.String(1), nullable=False)
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))


    __table_args__ = (sa.UniqueConstraint('label_id', 'resource_id', 'resource_name', 'resource_type', name='unique_label_resource'),)


class SchemaMigrations(Base):
    __tablename__ = 'schema_migrations'

    version = sa.Column(sa.BigInteger, primary_key=True)
    dirty = sa.Column(sa.Boolean, nullable=False)

class AdminJob(Base):
    __tablename__ = 'admin_job'

    id = sa.Column(sa.Integer, primary_key=True)
    job_name = sa.Column(sa.String(64), nullable=False)
    job_kind = sa.Column(sa.String(64), nullable=False)
    cron_str = sa.Column(sa.String(256))
    status = sa.Column(sa.String(64), nullable=False)
    job_uuid = sa.Column(sa.String(64))
    deleted = sa.Column(sa.Boolean, nullable=False, server_default='false')
    creation_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))
    update_time = sa.Column(sa.TIMESTAMP, server_default=sa.text("'now'::timestamp"))

    __table_args__ = (sa.Index('status', "job_uuid"),)