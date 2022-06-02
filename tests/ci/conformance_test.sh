#!/bin/bash
set -e

echo "get the conformance testing code..."
git clone https://github.com/opencontainers/distribution-spec.git

function createPro {
   echo "create testing project: $2"
   STATUS=$(curl -w '%{http_code}' -H 'Content-Type: application/json' -H 'Accept: application/json' -X POST -u "admin:Harbor12345" -s --insecure "https://$1/api/v2.0/projects" --data "{\"project_name\":\"$2\",\"metadata\":{\"public\":\"false\"},\"storage_limit\":-1}")
   if [ $STATUS -ne 201 ]; then
     echo "fail to create project: $2, rc: $STATUS"
     exit 1
   fi
}

createPro $1 conformance
createPro $1 crossmount

echo "run conformance test..."
export OCI_ROOT_URL="https://$1"
export OCI_NAMESPACE="conformance/testrepo"
export OCI_USERNAME="admin"
export OCI_PASSWORD="Harbor12345"
export OCI_DEBUG="true"

export OCI_TEST_PUSH=1
export OCI_TEST_PULL=1
export OCI_TEST_CONTENT_DISCOVERY=1
export OCI_TEST_CONTENT_MANAGEMENT=1
export OCI_CROSSMOUNT_NAMESPACE="crossmount/testrepo"
export OCI_AUTOMATIC_CROSSMOUNT="false"

cd ./distribution-spec/conformance
go test .