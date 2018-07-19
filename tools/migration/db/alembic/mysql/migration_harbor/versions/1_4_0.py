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

"""1.3.0 to 1.4.0

Revision ID: 1.3.0
Revises:

"""

# revision identifiers, used by Alembic.
revision = '1.4.0'
down_revision = '1.3.0'
branch_labels = None
depends_on = None

from alembic import op
from db_meta import *
import os

from sqlalchemy.dialects import mysql

Session = sessionmaker()

def upgrade():
    """
    update schema&data
    """
    bind = op.get_bind()
    session = Session(bind=bind)

    # Alter column length of project name
    op.alter_column('project', 'name', existing_type=sa.String(30), type_=sa.String(255), existing_nullable=False)

    # Add columns in replication_policy table
    op.add_column('replication_policy', sa.Column('filters', sa.String(1024)))
    op.add_column('replication_policy', sa.Column('replicate_deletion', mysql.TINYINT(1), nullable=False, server_default='0'))

    # create table replication_immediate_trigger
    ReplicationImmediateTrigger.__table__.create(bind)

    # Divided policies into unabled and enabled group
    unenabled_policies = session.query(ReplicationPolicy).filter(ReplicationPolicy.enabled == 0)
    enabled_policies = session.query(ReplicationPolicy).filter(ReplicationPolicy.enabled == 1)
    
    # As projects aren't stored in database of Harbor, migrate all replication
    # policies with manual trigger
    if os.getenv('WITH_ADMIRAL', '') == 'true':
        print ("deployed with admiral, migrating all replication policies with manual trigger")
        enabled_policies.update({
        ReplicationPolicy.enabled: 1,
        ReplicationPolicy.cron_str: '{"kind":"Manual"}'
    })
    else:
        # migrate enabeld policies
        enabled_policies.update({
            ReplicationPolicy.cron_str: '{"kind":"Immediate"}'
        })
        immediate_triggers = [ReplicationImmediateTrigger(
            policy_id=policy.id,
            namespace=session.query(Project).get(policy.project_id).name,
            on_push=1,
            on_deletion=1) for policy in enabled_policies]
        session.add_all(immediate_triggers)

    # migrate unenabled policies
    unenabled_policies.update({
        ReplicationPolicy.enabled: 1,
        ReplicationPolicy.cron_str: '{"kind":"Manual"}'
    })

    op.drop_constraint('PRIMARY', 'properties', type_='primary')
    op.create_unique_constraint('uq_properties_k', 'properties', ['k'])
    op.execute('ALTER TABLE properties ADD id INT PRIMARY KEY AUTO_INCREMENT;')

    session.commit()

def downgrade():
    """
    Downgrade has been disabled.
    """
