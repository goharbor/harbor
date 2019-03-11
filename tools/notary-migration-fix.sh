#!/bin/bash
# Copyright Project Harbor Authors
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

set -ex

harbor_db_image=$(docker images goharbor/harbor-db --format "{{.Repository}}:{{.Tag}}")
harbor_db_path="/data/database"

launch_db() {
    docker run -d --name fix-notary-migration -v ${harbor_db_path}:/var/lib/postgresql/data ${harbor_db_image} "postgres"
}

clean_db() {
    docker stop fix-notary-migration
    docker rm fix-notary-migration
}

wait_for_db_ready() {
    set +e
    TIMEOUT=12
    while [ $TIMEOUT -gt 0 ]; do
        docker exec fix-notary-migration pg_isready | grep "accepting connections"
        if [ $? -eq 0 ]; then
                break
        fi
        TIMEOUT=$((TIMEOUT - 1))
        sleep 5
    done
    if [ $TIMEOUT -eq 0 ]; then
        echo "Harbor DB cannot reach within one minute."
        clean_db
        exit 1
    fi
    set -e
}

fix_notary_server() {
    docker exec fix-notary-migration psql -U postgres -d notaryserver -c "delete from schema_migrations where version > 2 and not exists(select column_name from information_schema.columns where table_name = 'schema_migrations' and column_name = 'dirty');"
}

fix_notary_signer() {
    docker exec fix-notary-migration psql -U postgres -d notarysigner -c "delete from schema_migrations where version > 1 and not exists(select column_name from information_schema.columns where table_name = 'schema_migrations' and column_name = 'dirty');"
}


launch_db
wait_for_db_ready
fix_notary_server
fix_notary_signer
clean_db