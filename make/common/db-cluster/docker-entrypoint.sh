#!/bin/bash

# Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.
#
# This program is free software; you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation; version 2 of the License.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program; if not, write to the Free Software
# Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301 USA

CLUSTERCONF="/etc/mysql-cluster.cnf"
#MYCONF="/etc/my.cnf"
echo "" > $CLUSTERCONF
#echo "" > $MYCONF
IFS=$'\n'
COUNT=1
NDBCOUNT=0
for line in `cat /etc/hosts |grep db`
do
IFS=$' '
ary=(`echo $line | tr -s '	' ' '`)
if [ `echo ${ary[1]} | grep 'mgmd'` ] ; then
	echo -e "[ndb_mgmd]\nNodeId=$COUNT\nhostname=${ary[1]}\ndatadir=/var/lib/mysql\n" >> $CLUSTERCONF
elif [ `echo ${ary[1]} | grep 'ndbd'` ] ; then
	echo -e "[ndbd]\nNodeId=$COUNT\nhostname=${ary[1]}\ndatadir=/var/lib/mysql\nServerPort=50501\n"  >> $CLUSTERCONF
	NDBCOUNT=$(( NDBCOUNT + 1 ))
elif [ `echo ${ary[1]} | grep 'mysqld'` ] ; then
	echo -e "[mysqld]\nNodeId=$COUNT\nhostname=${ary[1]}\n\n"  >> $CLUSTERCONF
fi
COUNT=$(( COUNT + 1 ))
done

sed -i "1s/^/[ndbd default]\nNoOfReplicas=$NDBCOUNT\nDataMemory=80M\nIndexMemory=18M\n/" $CLUSTERCONF

set -e

ACTION="$1"
shift;

echo "Launching with user arguments: $@"

if [ $ACTION == "ndb_mgmd" ]; then
	echo "Starting ndb_mgmd"
	ndb_mgmd -f /etc/mysql-cluster.cnf --nodaemon "$@"

elif [ $ACTION == "ndbd" ]; then
	echo "Starting ndbd"
	ndbd --nodaemon "$@"

elif [ $ACTION == "mysqld" ]; then
	if [ ! -d '/var/lib/mysql/mysql' ]; then
		mysqld --user=mysql --initialize-insecure --datadir=/var/lib/mysql
		TEMP_FILE='/mysql-first-time.sql'
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
        	cat /r.sql >> "$TEMP_FILE"
		cat $TEMP_FILE

		chown -R mysql:mysql /var/lib/mysql
		mysqld --user=$MYSQL_USER --init-file="$TEMP_FILE"
	else
		mysqld --user=$MYSQL_USER
	fi
elif [ $ACTION == "ndb_mgm" ]; then
	echo "Starting ndb_mgm"
	ndb_mgm "$@"
else
	set -- "$ACTION" "$@"
	echo "Running custom user command $@"
	exec "$@"
fi
