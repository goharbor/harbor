# Copyright (c) 2008-2016 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""0.1.0 to 0.1.1

Revision ID: 0.1.1
Revises: 
Create Date: 2016-04-18 18:32:14.101897

"""

# revision identifiers, used by Alembic.
revision = '0.1.1'
down_revision = None
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import mysql
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime

Session = sessionmaker()

Base = declarative_base()

class Properties(Base):
    __tablename__ = 'properties'

    k = sa.Column(sa.String(64), primary_key = True)
    v = sa.Column(sa.String(128), nullable = False)

class ProjectMember(Base):
    __tablename__ = 'project_member'

    project_id = sa.Column(sa.Integer(), primary_key = True)
    user_id = sa.Column(sa.Integer(), primary_key = True)
    role = sa.Column(sa.Integer(), nullable = False)
    creation_time = sa.Column(sa.DateTime(), nullable = True)
    update_time = sa.Column(sa.DateTime(), nullable = True)
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

def upgrade():
    """
    update schema&data
    """
    bind = op.get_bind()
    session = Session(bind=bind)

    #delete M from table access
    acc = session.query(Access).filter_by(access_id=1).first()
    session.delete(acc)

    #create table property
    Properties.__table__.create(bind)
    session.add(Properties(k='schema_version', v='0.1.1'))

    #create table project_member
    ProjectMember.__table__.create(bind)

    #fill data
    join_result = session.query(UserProjectRole).join(UserProjectRole.project_role).all()
    for result in join_result:
        session.add(ProjectMember(project_id=result.project_role.project_id, \
            user_id=result.user_id, role=result.project_role.role_id, \
            creation_time=datetime.now(), update_time=datetime.now()))

    #drop user_project_role table before drop project_role
    #because foreign key constraint
    op.drop_table('user_project_role')
    op.drop_table('project_role')

    #add column to table project
    op.add_column('project', sa.Column('update_time', sa.DateTime(), nullable=True))

    #add column to table role
    op.add_column('role', sa.Column('role_mask', sa.Integer(), server_default=sa.text(u"'0'"), nullable=False))

    #add column to table user
    op.add_column('user', sa.Column('creation_time', sa.DateTime(), nullable=True))
    op.add_column('user', sa.Column('sysadmin_flag', sa.Integer(), nullable=True))
    op.add_column('user', sa.Column('update_time', sa.DateTime(), nullable=True))
    session.commit()

def downgrade():
    """
    Downgrade has been disabled.
    """
    pass
