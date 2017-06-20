#!/bin/bash
COMPOSE_CLUSTERTPL="./make/docker-compose-cluster.tpl"
THISNODEIP=$1
DB_BACKENDIP=$2
INITFLAG=$3
sed -n -e /version:\ \'2\'/,/tag:\ \"registry\"/p $COMPOSE_CLUSTERTPL
DBCLUSTERCONF="./make/db_cluster.cnf"
EXTRA_HOSTS=()
IFS=$'\n'
for line in `cat $DBCLUSTERCONF`
do
if [[ $line =~ \[.*\] ]] ; then
	NODE_IP=`echo $line | tr -d [ | tr -d ]`
	if [[ $NODE_IP = $THISNODEIP ]] ; then
		NODE_IP=$DB_BACKENDIP
	fi
else
	IFS=$' '
	ary=(`echo $line`)
	if [ `echo ${ary[0]} | grep 'mgmd'` ] ; then
		EXTRA_HOSTS=("${EXTRA_HOSTS[@]}" "db-mgmd-${ary[1]}:$NODE_IP")
	elif [ `echo ${ary[0]} | grep 'ndbd'` ] ; then
		EXTRA_HOSTS=("${EXTRA_HOSTS[@]}" "db-ndbd-${ary[1]}:$NODE_IP")
	elif [ `echo ${ary[0]} | grep 'mysqld'` ] ; then
		EXTRA_HOSTS=("${EXTRA_HOSTS[@]}" "db-mysqld-${ary[1]}:$NODE_IP")
	fi
fi
done

IFS=$'\n'
for line in `cat $DBCLUSTERCONF`
do
if [[ $line =~ \[.*\] ]] ; then
	NODE_IP=`echo $line | tr -d [ | tr -d ]`
	if [[ $NODE_IP = $THISNODEIP ]] ; then
		NODE_IP=$DB_BACKENDIP
	fi
fi
if [[ $NODE_IP = $DB_BACKENDIP ]] ; then
	IFS=$' '
	ary=(`echo $line`)
	SETIP=""
	if [ `echo ${ary[0]} | grep 'mgmd'` ] ; then
		echo -e "  mysql-management:\n    image: vmware/harbor-db:__version__\n    container_name: harbor-db-manegement\n    restart: always"
		echo -e "    extra_hosts:"
		for HOST in ${EXTRA_HOSTS[@]}; do
			if [[ $HOST =~ mgmd ]] ; then
				SETIP=$(echo $DB_BACKENDIP | sed -e s/\.[0-9]*/.2/4)
				echo -e "      - \"$HOST\"" | sed -e "s/$DB_BACKENDIP/$SETIP/g"
			else
				echo -e "      - \"$HOST\""
			fi
		done
		echo -e "    networks:\n      harbor:\n      db-backend:\n        ipv4_address: ${SETIP}\n"
		echo -e "    ports:\n      - 1186:1186"
		echo -e "    command: [\"/entrypoint.sh\", \"ndb_mgmd\"]"
	elif [ `echo ${ary[0]} | grep 'ndbd'` ] ; then
		echo -e "  mysql-ndb:\n    image: vmware/harbor-db:__version__\n    container_name: harbor-ndb\n    restart: always"
		echo -e "    extra_hosts:"
		for HOST in ${EXTRA_HOSTS[@]}; do
			if [[ $HOST =~ ndbd ]] ; then
				SETIP=$(echo $DB_BACKENDIP | sed -e s/\.[0-9]*/.3/4)
				echo -e "      - \"$HOST\"" | sed -e "s/$DB_BACKENDIP/$SETIP/g"
			else
				echo -e "      - \"$HOST\""
			fi
		done
		echo -e "    networks:\n      db-backend:\n        ipv4_address: ${SETIP}\n      harbor:"
		echo -e "    ports:\n      - 50501:50501"
		echo -e "    command: [\"/entrypoint.sh\", \"ndbd\"]"
	elif [ `echo ${ary[0]} | grep 'mysqld'` ] ; then
		echo -e "  mysql:\n    image: vmware/harbor-db:__version__\n    container_name: harbor-db\n    restart: always"
		echo -e "    volumes:\n      - /data/databases:/var/lib/mysql:z"
		echo -e "    env_file:\n      - ./common/config/db/env"
		if [ $INITFLAG = true ] ; then
			echo -e "    environment:\n      - INITFLAG=true"
		fi
		echo -e "    extra_hosts:"
		for HOST in ${EXTRA_HOSTS[@]}; do
			if [[ $HOST =~ mysqld ]] ; then
				SETIP=$(echo $DB_BACKENDIP | sed -e s/\.[0-9]*/.10/4)
				echo -e "      - \"$HOST\"" | sed -e "s/$DB_BACKENDIP/$SETIP/g"
			else
				echo -e "      - \"$HOST\""
			fi
		done
		echo -e "    networks:\n      db-backend:\n        ipv4_address: ${SETIP}\n      harbor:"
		echo -e "    ports:\n      - 3306:3306"
		echo -e "    depends_on:\n      - log\n    logging:\n      driver: \"syslog\"\n      options:\n        syslog-address: \"tcp://127.0.0.1:1514\"\n        tag: \"mysql\""
		echo -e "    command: [\"/entrypoint.sh\", \"mysqld\"]"

	fi
fi
done

sed -n -e /adminserver:/,/gateway:\ /p $COMPOSE_CLUSTERTPL
