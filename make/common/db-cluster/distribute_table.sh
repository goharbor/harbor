#!/bin/bash

MYSQL_PWD=`cat ./make/common/config/db/env |grep MYSQL_ROOT_PASSWORD | sed -e "s/MYSQL_ROOT_PASSWORD=//"`
MYSQL_IP=`docker network inspect root_db-backend | grep Gateway | sed  's/[^6][^0-9]*\([0-9.]*\)[^0-9]*.*/\1/'`
SQLFILE="./make/common/db-cluster/registry.sql"

IFS=$'\n'
#TABLES=`cat $SQLFILE |grep "create table" | cut -d" " -f3`
TABLES=`mysql -h ${MYSQL_IP} -u root -p${MYSQL_PWD} -e "show tables from registry;"`
for TABLE in ${TABLES}
do
	if [[ $TABLE = "Tables_in_registry" ]] ; then
		continue
	fi
	TABLE_CREATE=`mysql -h ${MYSQL_IP} -u root -p${MYSQL_PWD} -e "show create table registry.$TABLE;"`
	IFS=$' '
	for IBFK in ${TABLE_CREATE}
	do
		if [[ $IBFK =~ ibfk ]] ; then
			FK_NAME=$(echo $IBFK | tr -d \`)
			echo $FK_NAME
			mysql -h ${MYSQL_IP} -u root -p${MYSQL_PWD} -e  "alter table registry.$TABLE drop foreign key $FK_NAME;"
		fi
	done
done

IFS=$'\n'
for TABLE in ${TABLES}
do
	if [[ $TABLE = "Tables_in_registry" ]] ; then
		continue
	fi
	mysql -h ${MYSQL_IP} -u root -p${MYSQL_PWD} -e  "alter table registry.$TABLE engine ndbcluster;"
done

for LINE in `cat $SQLFILE`
do
	if [[ $LINE =~ "create table" ]] ; then
		TABLE=`echo $LINE | cut -d" " -f3`
	elif [[ $LINE =~ "FOREIGN KEY" ]] ; then
		LINE=$(echo $LINE | tr -d ,)
		mysql -h ${MYSQL_IP} -u root -p${MYSQL_PWD} -e  "alter table registry.$TABLE add $LINE;"
	fi
done
