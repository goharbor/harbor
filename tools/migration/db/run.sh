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
source $PWD/db/util/mysql.sh
source $PWD/db/util/pgsql.sh
source $PWD/db/util/mysql_pgsql_1_5_0.sh
source $PWD/db/util/alembic.sh

set -e

ISMYSQL=false
ISPGSQL=false
ISNOTARY=false
ISCLAIR=false

cur_version=""
PGSQL_USR="postgres" 

function init {
    if [ "$(ls -A /var/lib/mysql)" ]; then
        # As after the first success run, the data will be migrated to pgsql,
        # the PG_VERSION should be in /var/lib/mysql if user repeats the UP command.
        if [ -e '/var/lib/mysql/PG_VERSION' ]; then
            ISPGSQL=true
        elif [ -d '/var/lib/mysql/mysql' ]; then
            ISMYSQL=true
            if [ -d '/var/lib/mysql/notaryserver' ]; then
                ISNOTARY=true
            fi
        fi
    fi

    if [ "$(ls -A /var/lib/postgresql/data)" ]; then
        ISPGSQL=true
    fi

    if [ -d "/clair-db" ]; then
        ISCLAIR=true
    fi

    if [ $ISMYSQL == false ] && [ $ISPGSQL == false ]; then
        echo "No database has been mounted for the migration. Use '-v' to set it in 'docker run'."
        exit 1
    fi

    if [ $ISMYSQL == true ]; then
        # as for UP notary, user does not need to provide username and pwd.
        # the check works for harbor DB only.
        if [ $ISNOTARY == false ]; then
            if [ -z "$DB_USR" -o -z "$DB_PWD" ]; then
                echo "DB_USR or DB_PWD not set, exiting..."
                exit 1
            fi
            launch_mysql $DB_USR $DB_PWD
        else
            launch_mysql root
        fi
    fi

    if [ $ISPGSQL == true ]; then
        if [ $ISCLAIR == true ]; then
            launch_pgsql $PGSQL_USR "/clair-db"
        else
            launch_pgsql $PGSQL_USR
        fi
    fi
}

function get_version {
    if [ $ISMYSQL == true ]; then
        result=$(get_version_mysql) 
    fi
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
    if [ $ISMYSQL == true ]; then
        backup_mysql
    fi
    if [ $ISPGSQL == true ]; then
        backup_pgsql
    fi
    rc="$?"
    echo "Backup performed."
    exit $rc
}

function restore {
    echo "Performing restore..."
    if [ $ISMYSQL == true ]; then
        restore_mysql
    fi
    if [ $ISPGSQL == true ]; then
        restore_pgsql
    fi
    rc="$?"
    echo "Restore performed."
    exit $rc
}

function validate {
    echo "Performing test..."
    if [ $ISMYSQL == true ]; then
        test_mysql $DB_USR $DB_PWD
    fi
    if [ $ISPGSQL == true ]; then
        test_pgsql $PGSQL_USR
    fi
    rc="$?"
    echo "Test performed."
    exit $rc
}

function upgrade {
    if [ $ISNOTARY == true ];then
        up_notary $PGSQL_USR
    elif [ $ISCLAIR == true ];then
        up_clair $PGSQL_USR
    else
        up_harbor $1          
    fi   
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

    # $cur_version <='1.5.0', $target_version <='1.5.0', it needs to call mysql upgrade.
    if version_le $cur_version '1.5.0' && version_le $target_version '1.5.0'; then
        if [ $ISMYSQL != true ]; then
            echo "Please mount the database volume to /var/lib/mysql, then to run the upgrade again."
            return 1
        else
            alembic_up mysql $target_version
            return $?
        fi
    fi

    # $cur_version > '1.5.0', $target_version > '1.5.0', it needs to call pgsql upgrade.
    if ! version_le $cur_version '1.5.0' && ! version_le $target_version '1.5.0'; then    
        if [ $ISPGSQL != true ]; then
            echo "Please mount the database volume to /var/lib/postgresql/data, then to run the upgrade again."
            return 1
        else
            alembic_up pgsql $target_version
            return $?
        fi
    fi

    # $cur_version <='1.5.0', $target_version >'1.5.0', it needs to upgrade to $cur_version.mysql => 1.5.0.mysql => 1.5.0.pgsql => target_version.pgsql.
    if version_le $cur_version '1.5.0' && ! version_le $target_version '1.5.0'; then
        if [ $ISMYSQL != true ]; then
            echo "Please make sure to mount the correct the data volume."
            return 1
        else
            launch_pgsql $PGSQL_USR
            mysql_2_pgsql_1_5_0 $PGSQL_USR
            
            # Pgsql won't run the init scripts as the migration script has already created the PG_VERSION,
            # which is a flag that used by entrypoint.sh of pgsql to define whether to run init scripts to create harbor DBs.
            # Here to force init notary DBs just align with new harbor launch process.
            # Otherwise, user could get db failure when to launch harbor with notary as no data was created.

            psql -U $PGSQL_USR -f /harbor-migration/db/schema/notaryserver_init.pgsql
            psql -U $PGSQL_USR -f /harbor-migration/db/schema/notarysigner_init.pgsql

            ## it needs to call the alembic_up to target, disable it as it's now unsupported.
            #alembic_up $target_version
            stop_pgsql $PGSQL_USR
            stop_mysql $DB_USR $DB_PWD

            rm -rf /var/lib/mysql/*
            cp -rf $PGDATA/* /var/lib/mysql
            return 0
        fi        
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