#!/bin/bash

PGBINOLD=""
PGBINNEW="/usr/bin"
PGDATAOLD=""
PGDATANEW=""

while [[ "$#" -gt 0 ]]; do
        case $1 in
                -b|--old-datadir) PGDATAOLD="$2"; shift ;;
                -B|--new-datadir) PGDATANEW="$2"; shift ;;
                -d|--old-bindir) PGBINOLD="$2"; shift ;;
                -D|--new-bindir) PGBINNEW="$2"; shift ;;
                *) echo "Unknown parameter passed: $1"; exit 1 ;;
        esac
        shift
done

if [ "$PGDATAOLD" = "" ] || [ "$PGDATANEW" = "" ]; then
        echo "required parameter is missing: $PGDATAOLD, $PGDATANEW"
        exit 1
fi

export PGDATAOLD=$PGDATAOLD
export PGDATANEW=$PGDATANEW
export PGBINNEW=$PGBINNEW
export PGBINOLD=$PGBINOLD

echo 'start to upgrade.'
cd /tmp
${PGBINNEW}/pg_upgrade \
  --old-datadir=$PGDATAOLD \
  --new-datadir=$PGDATANEW \
  --old-bindir=$PGBINOLD \
  --new-bindir=$PGBINNEW \
  --old-options '-c config_file=$PGDATAOLD/postgresql.conf' \
  --new-options '-c config_file=$PGDATANEW/postgresql.conf'

if [ $? -ne 0 ]; then
        echo 'fail to upgrade.'
        cat /tmp/pg_upgrade_internal.log
        exit 1
fi

cp $PGDATAOLD/pg_hba.conf $PGDATANEW/pg_hba.conf

# Refresh collation version *metadata* on every connectable database to clear
# the "collation version mismatch" warnings that appear when the glibc version
# of the base image changes across PG major upgrades.
#
# NOTE: This updates metadata only. It does NOT rebuild indexes. Harbor's
# primary identifiers are expected to be ASCII, so the glibc collation-version
# change is not expected to affect normal Harbor ordering semantics.
#
# Operators with non-ASCII data or strict collation-order requirements should
# run `reindexdb -a -U postgres` during a maintenance window in addition to
# this metadata refresh.
echo 'refresh collation version metadata on all connectable databases.'

${PGBINNEW}/pg_ctl -D "$PGDATANEW" -w -o "-c listen_addresses='' -c unix_socket_directories='/run/postgresql'" start
if [ $? -ne 0 ]; then
        echo 'failed to start the new cluster for collation version refresh.'
        exit 1
fi

DBLIST=$(${PGBINNEW}/psql -h /run/postgresql -U postgres -d postgres -At -v ON_ERROR_STOP=1 -c "SELECT datname FROM pg_database WHERE datallowconn ORDER BY datname;")
if [ $? -ne 0 ]; then
        echo 'failed to list databases for collation version refresh.'
        ${PGBINNEW}/pg_ctl -D "$PGDATANEW" -m fast -w stop
        exit 1
fi

while IFS= read -r db; do
        [ -z "$db" ] && continue
        echo "  refreshing collation version for database: $db"
        ${PGBINNEW}/psql -h /run/postgresql -U postgres -d "$db" -v ON_ERROR_STOP=1 <<'SQL'
SELECT format('ALTER DATABASE %I REFRESH COLLATION VERSION;', current_database())
\gexec
SQL
        if [ $? -ne 0 ]; then
                echo "failed to refresh collation version for database: $db"
                ${PGBINNEW}/pg_ctl -D "$PGDATANEW" -m fast -w stop
                exit 1
        fi
done <<< "$DBLIST"

${PGBINNEW}/pg_ctl -D "$PGDATANEW" -m fast -w stop
if [ $? -ne 0 ]; then
        echo 'failed to stop the new cluster after collation version refresh.'
        exit 1
fi

echo 'success to upgrade.'