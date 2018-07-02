#!/usr/bin/env sh

# When run in the docker containers, the working directory
# is the root of the repo.

iter=0

case $SERVICE_NAME in
	notary_server)
		MIGRATIONS_PATH=${MIGRATIONS_PATH:-migrations/server/mysql}
		DB_URL=${DB_URL:-mysql://server@tcp(mysql:3306)/notaryserver}
		# have to poll for DB to come up
		until migrate -path=$MIGRATIONS_PATH -database="${DB_URL}" up
		do
			iter=$(( iter+1 ))
			if [[ $iter -gt 30 ]]; then
				echo "notaryserver database failed to come up within 30 seconds"
				exit 1;
			fi
			echo "waiting for $DB_URL to come up."
			sleep 1
		done
		echo "notaryserver database migrated to latest version"
		;;
	notary_signer)
		MIGRATIONS_PATH=${MIGRATIONS_PATH:-migrations/signer/mysql}
		DB_URL=${DB_URL:-mysql://signer@tcp(mysql:3306)/notarysigner}
		# have to poll for DB to come up
		until migrate -path=$MIGRATIONS_PATH -database="${DB_URL}" up
		do
			iter=$(( iter+1 ))
			if [[ $iter -gt 30 ]]; then
				echo "notarysigner database failed to come up within 30 seconds"
				exit 1;
			fi
			echo "waiting for $DB_URL to come up."
			sleep 1
		done
		echo "notarysigner database migrated to latest version"
		;;
esac
