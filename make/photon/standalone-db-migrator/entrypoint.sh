#!/bin/bash
set -e

pg_ctl start -w -t 60 -D ${PG_DATA}
/harbor/migrate
pg_ctl stop -D ${PG_DATA}