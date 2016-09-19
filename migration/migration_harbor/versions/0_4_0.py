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

"""0.3.0 to 0.4.0

Revision ID: 0.3.0
Revises: 

"""

# revision identifiers, used by Alembic.
revision = '0.4.0'
down_revision = '0.3.0'
branch_labels = None
depends_on = None

from alembic import op
from db_meta import *

from sqlalchemy.dialects import mysql

def upgrade():
    """
    update schema&data
    """
    bind = op.get_bind()
    #alter column user.username, alter column user.email, project.name and add column replication_policy.deleted
    op.alter_column('user', 'username', type_=sa.String(32), existing_type=sa.String(15))
    op.alter_column('user', 'email', type_=sa.String(255), existing_type=sa.String(128))
    op.alter_column('project', 'name', type_=sa.String(41), existing_type=sa.String(30), nullable=False)
    op.alter_column('replication_target', 'password', type_=sa.String(128), existing_type=sa.String(40))
    op.add_column('replication_policy', sa.Column('deleted', mysql.TINYINT(1), nullable=False, server_default=sa.text("'0'")))
    #create index pid_optime (project_id, op_time) on table access_log, poid_uptime (policy_id, update_time) on table replication_job
    op.create_index('pid_optime', 'access_log', ['project_id', 'op_time'])
    op.create_index('poid_uptime', 'replication_job', ['policy_id', 'update_time'])
    #create tables: repository
    Repository.__table__.create(bind)

def downgrade():
    """
    Downgrade has been disabled.
    """
    pass
