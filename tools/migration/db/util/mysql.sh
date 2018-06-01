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

DBCNF="-hlocalhost -u${DB_USR}"

function launch_mysql {
    set +e
    local usr="$1"
    local pwd="$2"
    if [ ! -z "$pwd" ]; then
        export MYSQL_PWD="${DB_PWD}"
    fi
    echo 'Trying to start mysql server...'
    chown -R 10000:10000 /var/lib/mysql
    mysqld &
    echo 'Waiting for MySQL start...'
    for i in {60..0}; do
        if [ -z "$pwd" ]; then
            mysqladmin -u$usr processlist >/dev/null 2>&1
        else
            mysqladmin -u$usr -p$pwd processlist >/dev/null 2>&1   
        fi
        if [ $? -eq 0 ]; then
            break
        fi
        sleep 1
    done
    set -e
    if [ "$i" -eq 0 ]; then
        echo "timeout. Can't run mysql server."
        return 1
    fi
    return 0
}

function test_mysql {
    set +e
    launch_mysql $DB_USR $DB_PWD
    if [ $? -eq 0 ]; then
        echo "DB test failed." 
        exit 0 
    else
        echo "DB test success." 
        exit 1
    fi
    set -e 
}

function stop_mysql {
    if [ -z $2 ]; then
        mysqladmin -u$1 shutdown
    else
        mysqladmin -u$1 -p$DB_PWD shutdown
    fi
    sleep 1
}

function get_version_mysql {
    local cur_version=""
    set +e
    if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
        where table_schema='registry' and table_name='alembic_version';") -eq 0 ]]; then
        echo "table alembic_version does not exist. Trying to initial alembic_version."
        mysql $DBCNF < ./alembic.sql
        #compatible with version 0.1.0 and 0.1.1
        if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
            where table_schema='registry' and table_name='properties'") -eq 0 ]]; then
            echo "table properties does not exist. The version of registry is 0.1.0"
            cur_version='0.1.0'
        else
            echo "The version of registry is 0.1.1"
            mysql $DBCNF -e "insert into registry.alembic_version values ('0.1.1')"
            cur_version='0.1.1'
        fi
    else
        cur_version=$(mysql $DBCNF -N -s -e "select * from registry.alembic_version;")
    fi
    set -e
    echo $cur_version
}

# It's only for registry, leverage the code from 1.5.0
function backup_mysql {
    mysqldump $DBCNF --add-drop-database --databases registry > /harbor-migration/backup/registry.sql
}

# It's only for registry, leverage the code from 1.5.0
function restore_mysql {
    mysql $DBCNF < /harbor-migration/backup/registry.sql
}
