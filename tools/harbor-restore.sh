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

extract_backup(){
    tar xvf harbor.tgz
}

launch_db() {
    if [ -n "$($DOCKER_CMD ps -q)" ]; then
        echo "There is running container, please stop and remove it before restore"
        exit 1
    fi
    $DOCKER_CMD run -d --name harbor-db -v ${PWD}:/backup -v ${harbor_db_path}:/var/lib/postgresql/data ${harbor_db_image} "postgres" 
}

clean_db() {
    $DOCKER_CMD stop harbor-db
    $DOCKER_CMD rm harbor-db
}

wait_for_db_ready() {
    set +e
    TIMEOUT=12
    while [ $TIMEOUT -gt 0 ]; do
        $DOCKER_CMD exec harbor-db pg_isready | grep "accepting connections"
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

clean_database_data(){
  set +e
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "drop database registry;" 
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "drop database postgres;"
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "drop database notarysigner; "
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "drop database notaryserver;"
  set -e 

  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "create database registry;"
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "create database postgres;"
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "create database notarysigner;"
  $DOCKER_CMD exec harbor-db psql -U postgres -d template1 -c "create database notaryserver;"
}

restore_database() {
    $DOCKER_CMD exec harbor-db sh -c 'psql -U postgres registry < /backup/harbor/db/registry.back'
    $DOCKER_CMD exec harbor-db sh -c 'psql -U postgres postgres < /backup/harbor/db/postgres.back'
    $DOCKER_CMD exec harbor-db sh -c 'psql -U postgres notarysigner < /backup/harbor/db/notarysigner.back'
    $DOCKER_CMD exec harbor-db sh -c 'psql -U postgres notaryserver < /backup/harbor/db/notaryserver.back'
}

restore_registry() {
    cp -r harbor/registry/* /data/registry
    chown -R 10000 /data/registry
}

restore_redis() {
    cp -r harbor/redis/* /data/redis
    chown -R 10000 /data/redis
}

restore_chartmuseum() {
    if [ -d ./harbor/chart_museum ]; then
        cp -r ./harbor/chart_museum/* /data/chart_museum
    fi
}

restore_secret() {
    if [ -f harbor/secret/secretkey ]; then
        cp -f harbor/secret/secretkey /data/secretkey 
    fi
    if [ -f harbor/secret/defaultalias ]; then
        cp -f harbor/secret/defaultalias /data/secretkey 
    fi
    if [ -d harbor/secret/keys ]; then
        cp -r harbor/secret/keys/ /data/secret/
    fi
}

note() { printf "\nNote:%s\n" "$@"
}

usage=$'harbor-restore.sh -- Backup Harbor script
./harbor-restore.sh   [options]          Restore Harbor with database and registry data      
Options: 
    --istile    Run restore in Harbor tile env
    --dbonly    Restore Harbor with database data only'

dbonly=false
istile=false
while [ $# -gt 0 ]; do
        case $1 in
            --help)
            note "$usage"
            exit 0;;
            --dbonly)
            dbonly=true;;
            --istile)
            istile=true;;
            *)
            note "$usage"
            exit 1;;
        esac
        shift || true
done

set -ex

if [ $istile = true ]; then
    DOCKER_CMD="/var/vcap/packages/docker/bin/docker -H unix:///var/vcap/sys/run/docker/dockerd.sock"
else
    DOCKER_CMD=docker
fi
harbor_db_image=$($DOCKER_CMD images goharbor/harbor-db --format "{{.Repository}}:{{.Tag}}" | head -1)
harbor_db_path="/data/database"

extract_backup
launch_db
wait_for_db_ready
clean_database_data
restore_database
restore_redis
if [ $dbonly = false ]; then
    restore_registry
    restore_chartmuseum
fi

restore_secret
clean_db
echo "All Harbor data is restored, you can start Harbor now"