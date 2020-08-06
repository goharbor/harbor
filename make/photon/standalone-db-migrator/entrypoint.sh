#!/bin/bash
set -e

pg_ctl start -w -t 60 -D ${PGDATA}
/harbor/migrate
pg_ctl stop -D ${PGDATA}