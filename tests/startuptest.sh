#!/bin/sh

set +e

TIMEOUT=12
while [ $TIMEOUT -gt 0 ]; do
    STATUS=$(curl --insecure -s -o /dev/null -w '%{http_code}' https://localhost/)
    if [ $STATUS -eq 200 ]; then
		break
    fi
    TIMEOUT=$(($TIMEOUT - 1))
    sleep 5
done

if [ $TIMEOUT -eq 0 ]; then
    echo "Harbor cannot reach within one minute."
    exit 1
fi

curl --insecure -s -L -H "Accept: application/json" https://localhost/ | grep "Harbor"  > /dev/null
if [ $? -eq 0 ]; then
	echo "Harbor is running success."
else
	echo "Harbor is running fail."
	exit 1
fi


