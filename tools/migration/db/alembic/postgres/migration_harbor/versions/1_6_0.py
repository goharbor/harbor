# Copyright Project Harbor Authors
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

"""1.5.0 to 1.6.0

Revision ID: 1.6.0
Revises:
Create Date: 2018-6-26

"""

# revision identifiers, used by Alembic.
revision = '1.6.0'
down_revision = '1.5.0'
branch_labels = None
depends_on = None

from alembic import op
from db_meta import *

Session = sessionmaker()

def upgrade():
    """
    update schema&data
    """
    bind = op.get_bind()
    session = Session(bind=bind)

    ## Add column deleted to harbor_label
    op.add_column('harbor_label', sa.Column('deleted', sa.Boolean, nullable=False, server_default='false'))

    ## Add schema_migration then insert data
    SchemaMigrations.__table__.create(bind)
    session.add(SchemaMigrations(version=1, dirty=False))

    ## Add table admin_job
    AdminJob.__table__.create(bind)
    op.execute('CREATE TRIGGER admin_job_update_time_at_modtime BEFORE UPDATE ON admin_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();')

    session.commit()

def downgrade():
    """
    Downgrade has been disabled.
    """
    pass
