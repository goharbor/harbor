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

"""1.2.0 to 1.3.0

Revision ID: 1.2.0
Revises:

"""

# revision identifiers, used by Alembic.
revision = '1.3.0'
down_revision = '1.2.0'
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

    # This is to solve the legacy issue when upgrade from 1.2.0rc1 to 1.3.0 refered by #3077
    username_coloumn = session.execute("show columns from user where field='username'").fetchone()
    if username_coloumn[1] != 'varchar(255)':
        op.alter_column('user', 'username', type_=sa.String(255))

    # create table project_metadata
    ProjectMetadata.__table__.create(bind)

    # migrate public data form project to project meta
    # The original type is int in project_meta data value type is string
    project_publicity = session.execute('SELECT project_id, public from project').fetchall()
    project_metadatas = [ProjectMetadata(project_id=project_id, name='public', value='true' if public else 'false')
                         for project_id, public in project_publicity]
    session.add_all(project_metadatas)

    # drop public column from project
    op.drop_column("project", "public")

    # add column insecure to replication target
    op.add_column('replication_target', sa.Column('insecure', mysql.TINYINT(1), nullable=False, server_default='0'))

    session.commit()

def downgrade():
    """
    Downgrade has been disabled.
    """
