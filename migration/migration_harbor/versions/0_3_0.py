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

"""0.1.1 to 0.3.0

Revision ID: 0.1.1
Revises: 

"""

# revision identifiers, used by Alembic.
revision = '0.3.0'
down_revision = '0.1.1'
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
    #alter column user.email, alter column access_log.repo_name, and add column access_log.repo_tag
    op.alter_column('user', 'email', type_=sa.String(128), existing_type=sa.String(30))
    op.alter_column('access_log', 'repo_name', type_=sa.String(256), existing_type=sa.String(40))
    try:
    	op.add_column('access_log', sa.Column('repo_tag', sa.String(128)))
    except Exception as e:
        if str(e).find("Duplicate column") >=0:
            print "ignore dup column error for repo_tag"
        else:
            raise e
    #create tables: replication_policy, replication_target, replication_job
    ReplicationPolicy.__table__.create(bind)
    ReplicationTarget.__table__.create(bind)
    ReplicationJob.__table__.create(bind)

def downgrade():
    """
    Downgrade has been disabled.
    """
    pass
