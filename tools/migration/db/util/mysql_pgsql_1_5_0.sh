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

DBCNF="-hlocalhost -u${DB_USR}"

function mysql_2_pgsql_1_5_0 {
    
    alembic_up mysql '1.5.0'

    ## dump 1.5.0-mysql
    mysqldump --compatible=postgresql --no-create-info --complete-insert --default-character-set=utf8 --databases registry > /harbor-migration/db/schema/registry.mysql

    ## migrate 1.5.0-mysql to 1.5.0-pqsql.
    python /harbor-migration/db/util/mysql_pgsql_data_converter.py /harbor-migration/db/schema/registry.mysql /harbor-migration/db/schema/registry_insert_data.pgsql

    ## import 1.5.0-pgsql into pgsql.
    psql -U $1 -f /harbor-migration/db/schema/registry_create_tables.pgsql
    psql -U $1 -f /harbor-migration/db/schema/registry_insert_data.pgsql

}

# This function is only for <= 1.5.0 to migrate notary db from mysql to pgsql.
function up_notary {

    set +e
    if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
                where table_schema='notaryserver' and table_name='tuf_files'") -eq 0 ]]; then
        echo "no content trust data needs to be updated."
        return 0
    else

        ## it's not a clean notary db, so cannot execute the create tables step.
        ## fail at here to call user to clean DB tables, then to run notary db migration.
        if [[ $(psql -U $1 -d notaryserver -t -c "select count(*) from pg_tables where schemaname='public';") -ne 0 ]]; then
            cat >&2 <<-EOF
                *******************************************************************************
                WARNING: Notary migration will only allow anyone haven't migrated notary or 
                         launched harbor yet. 
                         If you want to migrate notary data, please delete all the notaryserver 
                         and notarysigner DB tables in pgsql manually firstly.
                *******************************************************************************
EOF
            exit 0           
        fi

        set -e
        mysqldump --skip-triggers --compact --no-create-info --skip-quote-names --hex-blob --compatible=postgresql --default-character-set=utf8 --databases notaryserver > /harbor-migration/db/schema/notaryserver.mysql.tmp
        sed "s/0x\([0-9A-F]*\)/decode('\1','hex')/g" /harbor-migration/db/schema/notaryserver.mysql.tmp > /harbor-migration/db/schema/notaryserver_insert_data.mysql
        mysqldump --skip-triggers --compact --no-create-info --skip-quote-names --hex-blob --compatible=postgresql --default-character-set=utf8 --databases notarysigner > /harbor-migration/db/schema/notarysigner.mysql.tmp    
        sed "s/0x\([0-9A-F]*\)/decode('\1','hex')/g" /harbor-migration/db/schema/notarysigner.mysql.tmp > /harbor-migration/db/schema/notarysigner_insert_data.mysql

        python /harbor-migration/db/util/mysql_pgsql_data_converter.py /harbor-migration/db/schema/notaryserver_insert_data.mysql /harbor-migration/db/schema/notaryserver_insert_data.pgsql
        python /harbor-migration/db/util/mysql_pgsql_data_converter.py /harbor-migration/db/schema/notarysigner_insert_data.mysql /harbor-migration/db/schema/notarysigner_insert_data.pgsql

        # launch_pgsql $PGSQL_USR
        psql -U $1 -f /harbor-migration/db/schema/notaryserver_create_tables.pgsql
        psql -U $1 -f /harbor-migration/db/schema/notaryserver_insert_data.pgsql
        psql -U $1 -f /harbor-migration/db/schema/notaryserver_alter_tables.pgsql

        psql -U $1 -f /harbor-migration/db/schema/notarysigner_create_tables.pgsql
        psql -U $1 -f /harbor-migration/db/schema/notarysigner_insert_data.pgsql
        psql -U $1 -f /harbor-migration/db/schema/notarysigner_alter_tables.pgsql

        stop_mysql root
        stop_pgsql $1 
    fi
}

function up_clair {
    # clair DB info: user: 'postgres' database: 'postgres'

    set +e
    if [[ $(psql -U $1 -d postgres -t -c "select count(*) from vulnerability;") -eq 0 ]]; then
        echo "no vulnerability data needs to be updated."
        return 0
    else        
        pg_dump -U postgres postgres > /harbor-migration/db/schema/clair.pgsql
        stop_pgsql postgres "/clair-db"

        # it's harbor DB on pgsql.
        launch_pgsql $1
        ## it's not a clean clair db, so cannot execute the import step.
        ## fail at here to call user to clean DB, then to run clair db migration.
        if [[ $(psql -U $1 -d postgres -t -c "select count(*) from pg_tables where schemaname='public';") -ne 0 ]]; then
            cat >&2 <<-EOF
                *******************************************************************************
                WARNING: Clair migration will only allow anyone haven't migrated clair or 
                        launched harbor yet. 
                        If you want to migrate clair data, please delete all the clair DB tables 
                        in pgsql manually firstly.
                *******************************************************************************
EOF
            exit 0           
        fi
        set -e
        psql -U $1 -f /harbor-migration/db/schema/clair.pgsql
        stop_pgsql $1
    fi
}