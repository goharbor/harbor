#!/usr/bin/env sh

# When run in the docker containers, the working directory
# is the root of the repo.

iter=0

case $SERVICE_NAME in
	notary_server)
		MIGRATIONS_PATH=${MIGRATIONS_PATH:-migrations/server/mysql}
		DB_URL=${DB_URL:-mysql://server@tcp(mysql:3306)/notaryserver}
		# have to poll for DB to come up
		until migrate -path=$MIGRATIONS_PATH -url=$DB_URL version > /dev/null
		do
			iter=$(( iter+1 ))
			if [[ $iter -gt 30 ]]; then
				echo "notaryserver database failed to come up within 30 seconds"
				exit 1;
			fi
			echo "waiting for $DB_URL to come up."
			sleep 1
		done
		pre=$(migrate -path=$MIGRATIONS_PATH -url="${DB_URL}" version)
		if migrate -path=$MIGRATIONS_PATH -url="${DB_URL}" up ; then
			post=$(migrate -path=$MIGRATIONS_PATH -url="${DB_URL}" version)
			if [ "$pre" != "$post" ]; then
				echo "notaryserver database migrated to latest version"
			else
				echo "notaryserver database already at latest version"
			fi
		else
			echo "notaryserver database migration failed"
			exit 1
		fi
		;;
	notary_signer)
		MIGRATIONS_PATH=${MIGRATIONS_PATH:-migrations/signer/mysql}
		DB_URL=${DB_URL:-mysql://signer@tcp(mysql:3306)/notarysigner}
		# have to poll for DB to come up
		until migrate -path=$MIGRATIONS_PATH -url=$DB_URL up version > /dev/null
		do
			iter=$(( iter+1 ))
			if [[ $iter -gt 30 ]]; then
				echo "notarysigner database failed to come up within 30 seconds"
				exit 1;
			fi
			echo "waiting for $DB_URL to come up."
			sleep 1
		done
		pre=$(migrate -path=$MIGRATIONS_PATH -url="${DB_URL}" version)
		if migrate -path=$MIGRATIONS_PATH -url="${DB_URL}" up ; then
			post=$(migrate -path=$MIGRATIONS_PATH -url="${DB_URL}" version)
			if [ "$pre" != "$post" ]; then
				echo "notarysigner database migrated to latest version"
			else
				echo "notarysigner database already at latest version"
			fi
		else
			echo "notarysigner database migration failed"
			exit 1
		fi
		;;
esac
