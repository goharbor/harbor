#!/bin/sh

set +e

SWAGGER_ONLINE_VALIDATOR="http://online.swagger.io/validator"
if [ $TRAVIS_EVENT_TYPE = "push" ]; then
	HARBOR_SWAGGER_FILE="https://raw.githubusercontent.com/$TRAVIS_REPO_SLUG/$TRAVIS_COMMIT/api/harbor/swagger.yaml"
elif [ $TRAVIS_EVENT_TYPE = "pull_request" ]; then
	HARBOR_SWAGGER_FILE="https://raw.githubusercontent.com/$TRAVIS_PULL_REQUEST_SLUG/$TRAVIS_PULL_REQUEST_SHA/api/harbor/swagger.yaml"
else
	echo "* don't support this kinds of action ($TRAVIS_EVENT_TYPE), but don't fail the travis CI."
	exit 0
fi
HARBOR_SWAGGER_VALIDATOR_URL="$SWAGGER_ONLINE_VALIDATOR/debug?url=$HARBOR_SWAGGER_FILE"
echo $HARBOR_SWAGGER_VALIDATOR_URL

# Now try to ping swagger online validator, then to use it to do the validation.
eval curl -f -I $SWAGGER_ONLINE_VALIDATOR
curl_ping_res=$?
if [ ${curl_ping_res} -eq 0 ]; then
	echo "* cURL ping swagger validator returned success"
else
	echo "* cURL ping swagger validator returned an error (${curl_ping_res}), but don't fail the travis CI here."
	exit 0
fi

# Use the swagger online validator to validate the harbor swagger file.
eval curl -s $HARBOR_SWAGGER_VALIDATOR_URL > output.json
curl_validate_res=$?
validate_expected_results="{}"
validate_actual_results=$(cat < output.json)

if [ ${curl_validate_res} -eq 0 ]; then
	if [ $validate_actual_results = $validate_expected_results ]; then
		echo "* cURL check Harbor swagger file returned success"
	else
		echo "* cURL check Harbor swagger file returned an error ($validate_actual_results)"
	fi
else
	echo "* cURL check Harbor swagger file returned an error (${curl_validate_res})"
	exit ${curl_validate_res}
fi
