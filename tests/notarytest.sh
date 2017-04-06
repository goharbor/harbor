#!/bin/sh

set -e

TIMEOUT=10
while [ $TIMEOUT -gt 0 ]; do
    STATUS=$(curl -s -o /dev/null -w '%{http_code}' https://127.0.0.1:4443/v2/ -kv)
    if [ $STATUS -eq 401 ]; then
		echo "Notary is running success."
		break
    fi
    TIMEOUT=$(($TIMEOUT - 1))
    sleep 5
done

if [ $TIMEOUT -eq 0 ]; then
    echo "Notary is running fail."
    exit 1
fi
