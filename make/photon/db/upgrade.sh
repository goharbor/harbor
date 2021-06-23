#!/bin/bash

PGBINOLD="/usr/local/pg96/bin/"
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
echo 'success to upgrade.'