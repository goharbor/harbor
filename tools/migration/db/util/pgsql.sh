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

set -e

POSTGRES_PASSWORD=${DB_PWD}

function file_env {
        local var="$1"
        local fileVar="${var}_FILE"
        local def="${2:-}"
        if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
                echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
                exit 1
        fi
        local val="$def"
        if [ "${!var:-}" ]; then
                val="${!var}"
        elif [ "${!fileVar:-}" ]; then
                val="$(< "${!fileVar}")"
        fi
        export "$var"="$val"
        unset "$fileVar"
}

if [ "${1:0:1}" = '-' ]; then
        set -- postgres "$@"
fi

function launch_pgsql {
    local pg_data=$2
    if [ -z $2 ]; then
        pg_data=$PGDATA
    fi

    if [ "$1" = 'postgres' ]; then
            chown -R postgres:postgres $pg_data
            # look specifically for PG_VERSION, as it is expected in the DB dir
            if [ ! -s "$pg_data/PG_VERSION" ]; then
                    file_env 'POSTGRES_INITDB_ARGS'
                    if [ "$POSTGRES_INITDB_XLOGDIR" ]; then
                            export POSTGRES_INITDB_ARGS="$POSTGRES_INITDB_ARGS --xlogdir $POSTGRES_INITDB_XLOGDIR"
                    fi
                    su - $1 -c "initdb -D $pg_data  -U postgres -E UTF-8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 $POSTGRES_INITDB_ARGS"
                    # check password first so we can output the warning before postgres
                    # messes it up
                    file_env 'POSTGRES_PASSWORD'
                    if [ "$POSTGRES_PASSWORD" ]; then
                            pass="PASSWORD '$POSTGRES_PASSWORD'"
                            authMethod=md5
                    else
                            # The - option suppresses leading tabs but *not* spaces. :)
                            echo "Use \"-e POSTGRES_PASSWORD=password\" to set the password in \"docker run\"."
                            exit 1
                    fi

                    {
                            echo
                            echo "host all all all $authMethod"
                    } >> "$pg_data/pg_hba.conf"
                    # internal start of server in order to allow set-up using psql-client
                    # does not listen on external TCP/IP and waits until start finishes
                    su - $1 -c "pg_ctl -D \"$pg_data\" -o \"-c listen_addresses='localhost'\" -w start"

                    file_env 'POSTGRES_USER' 'postgres'
                    file_env 'POSTGRES_DB' "$POSTGRES_USER"

                    psql=( psql -v ON_ERROR_STOP=1 )

                    if [ "$POSTGRES_DB" != 'postgres' ]; then
                            "${psql[@]}" --username postgres <<-EOSQL
                                    CREATE DATABASE "$POSTGRES_DB" ;
EOSQL
                            echo
                    fi

                    if [ "$POSTGRES_USER" = 'postgres' ]; then
                            op='ALTER'
                    else
                            op='CREATE'
                    fi
                    "${psql[@]}" --username postgres <<-EOSQL
                            $op USER "$POSTGRES_USER" WITH SUPERUSER $pass ;
EOSQL
                    echo

                    psql+=( --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" )

                    echo
                    for f in /docker-entrypoint-initdb.d/*; do
                            case "$f" in
                                    *.sh)     echo "$0: running $f"; . "$f" ;;
                                    *.sql)    echo "$0: running $f"; "${psql[@]}" -f "$f"; echo ;;
                                    *.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
                                    *)        echo "$0: ignoring $f" ;;
                            esac
                            echo
                    done

                    #PGUSER="${PGUSER:-postgres}" \
                    #su - $1 -c "pg_ctl -D \"$pg_data\" -m fast -w stop"

                    echo
                    echo 'PostgreSQL init process complete; ready for start up.'
                    echo
            else
                su - $PGSQL_USR -c "pg_ctl -D \"$pg_data\" -o \"-c listen_addresses='localhost'\" -w start"
            fi
    fi
}

function stop_pgsql {
    local pg_data=$2
    if [ -z $2 ]; then
        pg_data=$PGDATA
    fi
    su - $1 -c "pg_ctl -D \"$pg_data\" -w stop"
}

function get_version_pgsql {
    version=$(psql -U $1 -d registry -t -c "select * from alembic_version;")
    echo $version
}

function test_pgsql {
    echo "TODO: needs to implement test pgsql connection..."
}

function backup_pgsql {
    echo "TODO: needs to implement backup registry..."
}

function restore_pgsql {
    echo "TODO: needs to implement restore registry..."
}
