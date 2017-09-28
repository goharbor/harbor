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

"""0.4.0 to 1.2.0

Revision ID: 0.4.0
Revises:

"""

# revision identifiers, used by Alembic.
revision = '1.2.0'
down_revision = '0.4.0'
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
	
    op.alter_column('user', 'realname', type_=sa.String(255), existing_type=sa.String(20))

    #delete column access_log.user_id(access_log_ibfk_1), access_log.project_id(access_log_ibfk_2)
    op.drop_constraint('access_log_ibfk_1', 'access_log', type_='foreignkey')
    op.drop_constraint('access_log_ibfk_2', 'access_log', type_='foreignkey')

	#add colume username to access_log
    op.add_column('access_log', sa.Column('username', mysql.VARCHAR(255), nullable=False))
    
	#init username
    session.query(AccessLog).update({AccessLog.username: ""})

    #update access_log username
    user_all = session.query(User).all()
    for user in user_all:
        session.query(AccessLog).filter(AccessLog.user_id == user.user_id).update({AccessLog.username: user.username}, synchronize_session='fetch')
	
    #update user.username length to 255
    op.alter_column('user', 'username', type_=sa.String(255), existing_type=sa.String(32))
	
    #update replication_target.username length to 255
    op.alter_column('replication_target', 'username', type_=sa.String(255), existing_type=sa.String(40))

    op.drop_column("access_log", "user_id")
    op.drop_column("repository", "owner_id")

    #create tables: img_scan_job, img_scan_overview, clair_vuln_timestamp
    ImageScanJob.__table__.create(bind)
    ImageScanOverview.__table__.create(bind)
    ClairVulnTimestamp.__table__.create(bind)

    session.commit()
		
def downgrade():
    """
    Downgrade has been disabled.
    """
    pass