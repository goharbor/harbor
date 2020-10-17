#!/bin/bash
set -e

[ $EXTERNAL_DB -eq 1 ] || pg_ctl start -w -t 60 -D ${PGDATA}
/harbor/migrate
[ $EXTERNAL_DB -eq 1 ] || pg_ctl stop -D ${PGDATA}
