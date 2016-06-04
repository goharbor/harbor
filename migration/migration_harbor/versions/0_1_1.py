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
from db_meta import *

from sqlalchemy.dialects import mysql

Session = sessionmaker()

def upgrade():
    """
    update schema&data
    """
    bind = op.get_bind()
    session = Session(bind=bind)

    #create table property
    Properties.__table__.create(bind)
    session.add(Properties(k='schema_version', v='0.1.1'))

    #add column to table user
    op.add_column('user', sa.Column('creation_time', mysql.TIMESTAMP, nullable=True))
    op.add_column('user', sa.Column('sysadmin_flag', sa.Integer(), nullable=True))
    op.add_column('user', sa.Column('update_time', mysql.TIMESTAMP, nullable=True))

    #init all sysadmin_flag = 0
    session.query(User).update({User.sysadmin_flag: 0})

    #create table project_member
    ProjectMember.__table__.create(bind)

    #fill data into project_member and user
    join_result = session.query(UserProjectRole).join(UserProjectRole.project_role).all()
    for result in join_result:
        session.add(ProjectMember(project_id=result.project_role.project_id, \
            user_id=result.user_id, role=result.project_role.role_id, \
            creation_time=None, update_time=None))

    #update sysadmin_flag
    sys_admin_result = session.query(UserProjectRole).\
        join(UserProjectRole.project_role).filter(ProjectRole.role_id ==1).all()
    for result in sys_admin_result:
        session.query(User).filter(User.user_id == result.user_id).update({User.sysadmin_flag: 1})

    #add column to table role
    op.add_column('role', sa.Column('role_mask', sa.Integer(), server_default=sa.text(u"'0'"), nullable=False))

    #drop user_project_role table before drop project_role
    #because foreign key constraint
    op.drop_table('user_project_role')
    op.drop_table('project_role')

    #delete sysadmin from table role
    role = session.query(Role).filter_by(role_id=1).first()
    session.delete(role)
    session.query(Role).update({Role.role_id: Role.role_id - 1})

    #delete A from table access
    acc = session.query(Access).filter_by(access_id=1).first()
    session.delete(acc)
    session.query(Access).update({Access.access_id: Access.access_id - 1})

    #add column to table project        
    op.add_column('project', sa.Column('update_time', mysql.TIMESTAMP, nullable=True))

    session.commit()

def downgrade():
    """
    Downgrade has been disabled.
    """
    pass
