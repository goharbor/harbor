#!/bin/sh

set +e

SWAGGERFILE=https://raw.githubusercontent.com/vmware/harbor/$1/docs/swagger.yaml
VALIDATOR=http://online.swagger.io/validator/debug?url=$SWAGGERFILE

echo $SWAGGERFILE

TIMEOUT=10
while [ $TIMEOUT -gt 0 ]; do
    STATUS=$(curl --insecure -s -o /dev/null -w '%{http_code}' $VALIDATOR)
    if [ $STATUS -eq 200 ]; then
		break
    fi
    TIMEOUT=$(($TIMEOUT - 1))
    sleep 2
done

if [ $TIMEOUT -eq 0 ]; then
    echo "Swagger online checke cannot reach, would not fail travis here."
    exit 0
fi

curl -X GET $VALIDATOR | grep "{}"  > /dev/null
if [ $? -eq 0 ]; then 
	echo "Swagger yaml check success."
else
	echo "Swagger yaml check fail."
	echo $(curl -X GET $VALIDATOR)
	exit 1
fi
 