#!/bin/bash
set -e

source $PWD/initdb.sh

CUR=$PWD
PG_VERSION_OLD=$1
PG_VERSION_NEW=$2

PGBINOLD="/usr/pgsql/${PG_VERSION_OLD}/bin"

PGDATAOLD=${PGDATA}/pg${PG_VERSION_OLD}
PGDATANEW=${PGDATA}/pg${PG_VERSION_NEW}

# We should block the upgrade path from 9.6 directly.
if [ -s $PGDATA/PG_VERSION ]; then
        echo "Upgrading from PostgreSQL 9.6 to PostgreSQL $PG_VERSION_NEW is not supported in the current Harbor release."
        echo "You should upgrade to previous Harbor firstly, then upgrade to current release."
        exit 1
fi

# Upgrade DB: 1. PG_NEW\PG_VERSION file doesn’t exist and pg_old_parameter is not nil and PG_OLD\PG_VERSION file exist.
#             For example: ["15", "18"]
#             In harbor v2.13, Harbor 2.12 was installed before, db version was 15,
#             It needs to upgrade the database from pg 15 to pg 18,
#             ["15", "18"] means support for upgrading from pg 15 to pg 18.
# Init DB:    1. PG_NEW\PG_VERSION file doesn’t exist and pg_old_parameter is not nil and PG_OLD\PG_VERSION file doesn’t exist.
#             For example: ["15", "18"]
#             In harbor v2.13, the first time installation, it needs to init the db for pg 18,
#             ["15", "18"] means support for upgrading from pg 15 to pg 18.
#             2. PG_NEW\PG_VERSION file doesn’t exist and pg_old_parameter is nil.
#             For example: ["", "18"]
#             In harbor v2.13, the first time installation, it needs to init the db for pg 18,
#             ["", "18"] means db upgrade is not supported.
if [ ! -s $PGDATANEW/PG_VERSION ]; then
        if [ ! -z $PG_VERSION_OLD ] && [ -s $PGDATAOLD/PG_VERSION ]; then
                echo "upgrade DB from $PG_VERSION_OLD to $PG_VERSION_NEW"
                # Match the data checksum setting of the old cluster, otherwise pg_upgrade
                # will fail with "old cluster does not use data checksums but the new one does".
                # Starting from PostgreSQL 18, initdb enables data checksums by default while
                # earlier versions (e.g. PG 15) do not. Detect the old cluster's checksum
                # state via pg_controldata and pass --no-data-checksums to initdb when needed.
                if ! OLD_CONTROLDATA=$($PGBINOLD/pg_controldata "$PGDATAOLD" 2>/dev/null); then
                        echo "failed to determine old cluster checksum setting: pg_controldata failed"
                        exit 1
                fi
                # Parse the "Data page checksum version" line with bash's built-in
                # regex match so we don't depend on awk/sed/grep.
                if [[ "$OLD_CONTROLDATA" =~ Data\ page\ checksum\ version:[[:space:]]*([0-9]+) ]]; then
                        OLD_CHECKSUM_VERSION="${BASH_REMATCH[1]}"
                else
                        echo "failed to determine old cluster checksum setting: unexpected pg_controldata output"
                        exit 1
                fi
                if [ "$OLD_CHECKSUM_VERSION" = "0" ]; then
                        echo "old cluster has data checksums disabled, initializing new cluster with --no-data-checksums"
                        export POSTGRES_INITDB_ARGS="${POSTGRES_INITDB_ARGS} --no-data-checksums"
                fi
                initPG $PGDATANEW false
                set +e
                # In some cases, like helm upgrade, the postgresql may not quit cleanly.
                # Use start & stop to clean the unexpected status. Error:
                #   There seems to be a postmaster servicing the new cluster.
                #   Please shutdown that postmaster and try again.
                #   Failure, exiting
                $PGBINOLD/pg_ctl -D "$PGDATAOLD" -w -o "-p 5433" start
                $PGBINOLD/pg_ctl -D "$PGDATAOLD" -m fast -w stop
                ./$CUR/upgrade.sh --old-bindir $PGBINOLD --old-datadir $PGDATAOLD --new-datadir $PGDATANEW
                # it needs to clean the $PGDATANEW on upgrade failure
                if [ $? -ne 0 ]; then
                        echo "remove the $PGDATANEW after fail to upgrade."
                        rm -rf $PGDATANEW
                        exit 1
                fi
                set -e
                echo "remove the $PGDATAOLD after upgrade success."
                rm -rf $PGDATAOLD
        else
                echo "init DB, DB version:$PG_VERSION_NEW"
                initPG $PGDATANEW true
        fi
fi

POSTGRES_PARAMETER=''
file_env 'POSTGRES_MAX_CONNECTIONS' '1024'
# The max value of 'max_connections' is 262143
if [ $POSTGRES_MAX_CONNECTIONS -le 0 ] || [ $POSTGRES_MAX_CONNECTIONS -gt 262143 ]; then
        POSTGRES_MAX_CONNECTIONS=262143
fi

POSTGRES_PARAMETER="${POSTGRES_PARAMETER} -c max_connections=${POSTGRES_MAX_CONNECTIONS}"
exec postgres -D $PGDATANEW $POSTGRES_PARAMETER
