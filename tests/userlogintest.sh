#!/bin/sh

set +e

STATUS_LOGIN=$(curl --insecure -w '%{http_code}' -d "principal=$1&password=$2" https://localhost/c/login)
if [ $STATUS_LOGIN -eq 200 ]; then
	echo "Login Harbor success."
else
	echo "Login Harbor fail."
	exit 1
fi


STATUS_LOGOUT=$(curl --insecure -s -o /dev/null -w '%{http_code}' https://localhost/c/log_out)
if [ $STATUS_LOGOUT -eq 200 ]; then
	echo "Logout Harbor success."
else
	echo "Logout Harbor fail."
	exit 1
fi
