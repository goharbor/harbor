#!/bin/bash
set +e
if [ ! -f /var/lib/mysql/created_in_mariadb.flag ]; then
    echo "Maria DB flag not found, the DB was created in mysql image, running upgrade..."
    mysqld >/dev/null 2>&1 &
    pid="$!"
    for i in {30..0}; do
        mysqladmin -uroot -p$MYSQL_ROOT_PASSWORD processlist >/dev/null 2>&1
        if [ $? = 0 ]; then
            break
        fi
        echo 'Waiting for MySQL start...'
        sleep 1
    done
    if [ "$i" = 0 ]; then
        echo >&2 'MySQL failed to start.'
        exit 1
    fi
    set -e
    mysql_upgrade -p$MYSQL_ROOT_PASSWORD
    echo 'Finished upgrading'
    if ! kill -s TERM "$pid" || ! wait "$pid"; then
        echo >&2 'Failed to stop MySQL for upgrading.'
        exit 1
    fi
else
    echo "DB was created in Maria DB, skip upgrade."
fi
