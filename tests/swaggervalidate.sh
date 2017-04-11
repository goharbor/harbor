#!/bin/sh

set +e

if [ -z "$1" ]; then
	echo '* Required input `git branch` not provided!'
	exit 1
fi

SWAGGER_VALIDATOR="http://online.swagger.io/validator"
SWAGGER_FILE="https://raw.githubusercontent.com/vmware/harbor/$1/docs/swagger.yaml"
VALIDATOR="$SWAGGER_VALIDATOR/debug?url=$SWAGGER_FILE"
echo $SWAGGER_FILE

# Now try to validate swagger online validator, then to use it to do the validation.
eval curl -f -I $SWAGGER_VALIDATOR
curl_ping_res=$?
if [ ${curl_ping_res} -eq 0 ]; then
	echo "* cURL ping swagger validator returned success"
else
	echo "* cURL ping swagger validator returned an error (${curl_ping_res})"
	exit ${curl_ping_res}
fi

# Use the validator to validate the swagger file.
eval curl -s $VALIDATOR > output.json
curl_validate_res=$?
validate_expected_results="{}"
validate_actual_results=$(cat < output.json)

if [ ${curl_ping_res} -eq 0 ]; then
	if [ $validate_actual_results = $validate_expected_results ]; then
		echo "* cURL check swagger file returned success"
	else
		echo "* cURL check swagger file returned an error ($validate_actual_results)"
	fi
else
	echo "* cURL check swagger file returned an error (${curl_validate_res})"
	exit ${curl_validate_res}
fi
