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

create_dir(){
    rm -rf harbor
    mkdir -p harbor/db
    mkdir -p harbor/secret
}

launch_db() {
    if [ -n "$($DOCKER_CMD ps -q)" ]; then
        echo "There is running container, please stop and remove it before backup"
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

dump_database() {
    $DOCKER_CMD exec harbor-db sh -c 'pg_dump -U postgres registry > /backup/harbor/db/registry.back'
    $DOCKER_CMD exec harbor-db sh -c 'pg_dump -U postgres postgres > /backup/harbor/db/postgres.back'
    $DOCKER_CMD exec harbor-db sh -c 'pg_dump -U postgres notarysigner > /backup/harbor/db/notarysigner.back'
    $DOCKER_CMD exec harbor-db sh -c 'pg_dump -U postgres notaryserver > /backup/harbor/db/notaryserver.back'
}

backup_registry() {
    cp -rf /data/registry  harbor/
}

backup_chart_museum() {
    if [ -d /data/chart_storage ]; then
        cp -rf /data/chart_storage harbor/
    fi
}

backup_redis() {
    if [ -d /data/redis ]; then
        cp -rf /data/redis harbor/
    fi
}

backup_secret() {
    if [ -f /data/secretkey ]; then
        cp /data/secretkey harbor/secret/
    fi
    if [ -f /data/defaultalias ]; then
         cp /data/defaultalias harbor/secret/
    fi
    # location changed after 1.8.0
    if [ -d /data/secret/keys/ ]; then
        cp -r /data/secret/keys/ harbor/secret/
    fi
}

create_tarball() {
    tar zcvf harbor.tgz harbor
    rm -rf harbor
}

note() { printf "\nNote:%s\n" "$@"
}

usage=$'harbor-backup.sh -- Backup Harbor script
./harbor-backup.sh      [options]   Backup Harbor with database and registry data      
Options
    --istile    Backup in Harbor tile env
    --dbonly    Backup Harbor with database data only
'
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
harbor_db_image=$($DOCKER_CMD images goharbor/harbor-db --format "{{.Repository}}:{{.Tag}}" |head -1)
harbor_db_path="/data/database"


create_dir
launch_db
wait_for_db_ready
dump_database
backup_redis
if [ $dbonly = false ];  then
    backup_registry
    backup_chart_museum
fi
backup_secret
create_tarball
clean_db

echo "All Harbor data are backed up"