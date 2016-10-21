#!/bin/bash
set -e

if [ ! -d '/var/lib/mysql/mysql' -a "${1%_safe}" = 'mysqld' ]; then
	if [ -z "$MYSQL_ROOT_PASSWORD" -a -z "$MYSQL_ALLOW_EMPTY_PASSWORD" ]; then
		echo >&2 'error: database is uninitialized and MYSQL_ROOT_PASSWORD not set'
		echo >&2 '  Did you forget to add -e MYSQL_ROOT_PASSWORD=... ? v2'
		exit 1
	fi
	
	mysql_install_db --user=mysql --datadir=/var/lib/mysql
	
	# These statements _must_ be on individual lines, and _must_ end with
	# semicolons (no line breaks or comments are permitted).
	# TODO proper SQL escaping on ALL the things D:
	TEMP_FILE='/tmp/mysql-first-time.sql'
	cat > "$TEMP_FILE" <<-EOSQL
		DELETE FROM mysql.user ;
		CREATE USER 'root'@'%' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}' ;
		GRANT ALL ON *.* TO 'root'@'%' WITH GRANT OPTION ;
		DROP DATABASE IF EXISTS test ;
	EOSQL
	
	if [ "$MYSQL_DATABASE" ]; then
		echo "CREATE DATABASE IF NOT EXISTS $MYSQL_DATABASE ;" >> "$TEMP_FILE"
	fi
	
	if [ "$MYSQL_USER" -a "$MYSQL_PASSWORD" ]; then
		echo "CREATE USER '$MYSQL_USER'@'%' IDENTIFIED BY '$MYSQL_PASSWORD' ;" >> "$TEMP_FILE"
		
		if [ "$MYSQL_DATABASE" ]; then
			echo "GRANT ALL ON $MYSQL_DATABASE.* TO '$MYSQL_USER'@'%' ;" >> "$TEMP_FILE"
		fi
	fi
	
	echo 'FLUSH PRIVILEGES ;' >> "$TEMP_FILE"
	cat /tmp/r.sql >> "$TEMP_FILE"
	
	set -- "$@" --init-file="$TEMP_FILE"
fi

chown -R mysql:mysql /var/lib/mysql
exec "$@"
