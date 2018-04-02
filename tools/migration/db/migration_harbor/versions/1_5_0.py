# Copyright (c) 2008-2018 VMware, Inc. All Rights Reserved.
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

"""1.4.0 to 1.5.0

Revision ID: 1.5.0
Revises:

"""

# revision identifiers, used by Alembic.
revision = '1.5.0'
down_revision = '1.4.0'
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


    # create table harbor_label
    HarborLabel.__table__.create(bind)

    # create table harbor_resource_label
    HarborResourceLabel.__table__.create(bind)

    # create user_group
    UserGroup.__table__.create(bind)

    # project member
    op.drop_constraint('project_member_ibfk_1', 'project_member', type_='foreignkey')
    op.drop_constraint('project_member_ibfk_2', 'project_member', type_='foreignkey')
    op.drop_constraint('project_member_ibfk_3', 'project_member', type_='foreignkey')
    op.drop_constraint('PRIMARY', 'project_member', type_='primary')
    op.drop_index('user_id', 'project_member')
    op.drop_index('role', 'project_member')
    op.execute('ALTER TABLE project_member ADD id INT PRIMARY KEY AUTO_INCREMENT;')
    op.alter_column('project_member', 'user_id', existing_type=sa.Integer, existing_nullable=False, new_column_name='entity_id')
    op.alter_column('project_member', 'creation_time', existing_type=mysql.TIMESTAMP, server_default = sa.text("CURRENT_TIMESTAMP"))
    op.alter_column('project_member', 'update_time', existing_type=mysql.TIMESTAMP, server_default=sa.text("CURRENT_TIMESTAMP"), onupdate=sa.text("CURRENT_TIMESTAMP"))
    op.add_column('project_member', sa.Column('entity_type', sa.String(1)))

    session.query(ProjectMember).update({
        ProjectMember.entity_type: 'u'
    })
    op.alter_column('project_member', 'entity_type', existing_type=sa.String(1), existing_nullable=True, nullable=False)

    op.create_unique_constraint('unique_project_entity_type', 'project_member', ['project_id', 'entity_id', 'entity_type'])

    # add job_uuid to replicationjob and img_scan_job
    op.add_column('replication_job', sa.Column('job_uuid', sa.String(64)))
    op.add_column('img_scan_job', sa.Column('job_uuid', sa.String(64)))

    # add index to replication job
    op.create_index('poid_status', 'replication_job', ['policy_id', 'status'])

    # add index to img_scan_job
    op.create_index('idx_status', 'img_scan_job', ['status'])
    op.create_index('idx_digest', 'img_scan_job', ['digest'])
    op.create_index('idx_uuid', 'img_scan_job', ['job_uuid'])
    op.create_index('idx_repository_tag', 'img_scan_job', ['repository', 'tag'])

    session.commit()

def downgrade():
    """
    Downgrade has been disabled.
    """
