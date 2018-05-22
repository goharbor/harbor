#!/bin/bash
# Copyright 2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

function alembic_up {
    local db_type="$1"
    local target_version="$2"

    export PYTHONPATH=$PYTHONPATH:/harbor-migration/db/alembic
    
    if [ $db_type = "mysql" ]; then
        source /harbor-migration/db/alembic/alembic.tpl > /harbor-migration/db/alembic/alembic.ini     
        echo "Performing upgrade $target_version..."
        alembic -c /harbor-migration/db/alembic/alembic.ini current
        alembic -c /harbor-migration/db/alembic/alembic.ini upgrade $target_version
        alembic -c /harbor-migration/db/alembic/alembic.ini current
    elif [ $db_type = "pgsql" ]; then
        echo "TODO: add support for pgsql."
    else
        echo "Unsupported DB type."
        exit 1
    fi

    echo "Upgrade performed."
}