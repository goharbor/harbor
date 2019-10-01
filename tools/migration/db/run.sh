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
source $PWD/db/util/pgsql.sh
source $PWD/db/util/alembic.sh

set -e

ISPGSQL=false

cur_version=""
PGSQL_USR="postgres"

function init {
    if [ "$(ls -A /var/lib/postgresql/data)" ]; then
        ISPGSQL=true
    fi

    if  [ $ISPGSQL == false ]; then
        echo "No database has been mounted for the migration. Use '-v' to set it in 'docker run'."
        exit 1
    fi

    if [ $ISPGSQL == true ]; then
        launch_pgsql $PGSQL_USR
    fi
}

function get_version {
    if [ $ISPGSQL == true ]; then
        result=$(get_version_pgsql $PGSQL_USR)
    fi
    cur_version=$result
}

# first version is less than or equal to second version.
function version_le {
    ## if no version specific, see it as larger then 1.5.0
    if [ $1 = "head" ];then
        return 1
    fi
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" = "$1";
}

function backup {
    echo "Performing backup..."
    if [ $ISPGSQL == true ]; then
        backup_pgsql
    fi
    rc="$?"
    echo "Backup performed."
    exit $rc
}

function restore {
    echo "Performing restore..."
    if [ $ISPGSQL == true ]; then
        restore_pgsql
    fi
    rc="$?"
    echo "Restore performed."
    exit $rc
}

function validate {
    echo "Performing test..."
    if [ $ISPGSQL == true ]; then
        test_pgsql $PGSQL_USR
    fi
    rc="$?"
    echo "Test performed."
    exit $rc
}

function upgrade {
    up_harbor $1
}

function up_harbor {
    local target_version="$1"
    if [ -z $target_version ]; then
        target_version="head"
        echo "Version is not specified. Default version is head."
    fi

    get_version
    if [ "$cur_version" = "$target_version" ]; then
        echo "It has always running the $target_version, no longer need to upgrade."
        exit 0
    fi

    # $cur_version > '1.5.0', $target_version > '1.5.0', it needs to call pgsql upgrade.
    if [ $ISPGSQL != true ]; then
        echo "Please mount the database volume to /var/lib/postgresql/data, then to run the upgrade again."
        return 1
    else
        alembic_up pgsql $target_version
        return $?
    fi

    echo "Unsupported DB upgrade from $cur_version to $target_version, please check the inputs."
    return 1
}

function main {

    if [[ $1 = "help" || $1 = "h" || $# = 0 ]]; then
        echo "Usage:"
        echo "backup                perform database backup"
        echo "restore               perform database restore"
        echo "up,   upgrade         perform database schema upgrade"
        echo "test                  test database connection"
        echo "h,    help            usage help"
        exit 0
    fi

    init

    local key="$1"

    case $key in
    up|upgrade)
        upgrade $2
        ;;
    backup)
       backup
        ;;
    restore)
       restore
        ;;
    test)
       validate
        ;;
    *)
        echo "unknown option"
        exit 0
        ;;
    esac
}

main "$@"
