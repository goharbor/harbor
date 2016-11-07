#!/bin/bash

export PYTHONPATH=$PYTHONPATH:/harbor-migration
if [ -z "$DB_USR" -o -z "$DB_PWD" ]; then
    echo "DB_USR or DB_PWD not set, exiting..."
    exit 1
fi

source ./alembic.tpl > ./alembic.ini

WAITTIME=60

DBCNF="-hlocalhost -u${DB_USR}"

#prevent shell to print insecure message
export MYSQL_PWD="${DB_PWD}"

if [[ $1 = "help" || $1 = "h" || $# = 0 ]]; then
    echo "Usage:"
    echo "backup                perform database backup"
    echo "restore               perform database restore"
    echo "up,   upgrade         perform database schema upgrade"
    echo "test                  test database connection"
    echo "h,    help            usage help"
    exit 0
fi

if [[ ( $1 = "up" || $1 = "upgrade" ) && ${SKIP_CONFIRM} != "y" ]]; then
    echo "Please backup before upgrade."
    read -p "Enter y to continue updating or n to abort:" ans
    case $ans in
        [Yy]* )
            ;;
        [Nn]* ) 
            exit 0
            ;;
        * ) echo "illegal answer: $ans. Upgrade abort!!"
            exit 1
            ;;
    esac

fi

echo 'Trying to start mysql server...'
DBRUN=0
mysqld &
for i in $(seq 1 $WAITTIME); do
    echo "$(/usr/sbin/service mysql status)"
    if [[ "$(/usr/sbin/service mysql status)" =~ "not running" ]]; then
        sleep 1
    else
        DBRUN=1
        break
    fi
done

if [[ $DBRUN -eq 0  ]]; then
    echo "timeout. Can't run mysql server."
    if [[ $1 = "test" ]]; then
        echo "test failed."
    fi
    exit 1
fi

if [[ $1 = "test" ]]; then
    echo "test passed."
    exit 0
fi

key="$1"
case $key in
up|upgrade)
    VERSION="$2"
    if [[ -z $VERSION ]]; then
        VERSION="head"
        echo "Version is not specified. Default version is head."
    fi
    echo "Performing upgrade ${VERSION}..."
    if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
        where table_schema='registry' and table_name='alembic_version';") -eq 0 ]]; then
        echo "table alembic_version does not exist. Trying to initial alembic_version."
        mysql $DBCNF < ./alembic.sql
        #compatible with version 0.1.0 and 0.1.1
        if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
            where table_schema='registry' and table_name='properties'") -eq 0 ]]; then
            echo "table properties does not exist. The version of registry is 0.1.0"
        else
            echo "The version of registry is 0.1.1"
            mysql $DBCNF -e "insert into registry.alembic_version values ('0.1.1')"
        fi
    fi
    alembic -c ./alembic.ini upgrade ${VERSION}
    echo "Upgrade performed."
    ;;
backup)
    echo "Performing backup..."
    mysqldump $DBCNF --add-drop-database --databases registry > ./backup/registry.sql
    echo "Backup performed."
    ;;
restore)
    echo "Performing restore..."
    mysql $DBCNF < ./backup/registry.sql
    echo "Restore performed."
    ;;
*)
    echo "unknown option"
    exit 0
    ;;
esac
