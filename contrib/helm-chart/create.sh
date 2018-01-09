#!/usr/bin/env bash
set -ex

SECURITY_DIR="/path/to/secrets/directory"
# SSL_CERT=${SECURITY_DIR}/SSL_CERT.crt
NAMESPACE="harbor"
PROVIDER=${1};
REGION=${2};
ISO_3166_COUNTRY_CODE="DE"
STATE="Berlin"
CITY="Berlin"
GCS_KEYFILE_PATH=${3};
SUBDOMAIN="foo.bar.example.com"
CLUSTER_ID="k8s-${REGION}-${PROVIDER}"
CLUSTER_SECRETS_DIR="${SECURITY_DIR}/k8s/${CLUSTER_ID}/${NAMESPACE}"
EXTERNAL_REGISTRY_URL="registry-${REGION}-${PROVIDER}.${SUBDOMAIN}"

# example of starting this process
# ./create.sh azure ue1
# ./create.sh aws uw2
# ./create.sh gcp uw1 /path/to/gcs_keyfile

mkdir -p ${CLUSTER_SECRETS_DIR}
# NOTE: If you're using Google Cloud Storage for your registry data storage
#       make sure you've created and supplied the path to properly formed
#       gcs_keyfile
if [ $PROVIDER = "gcp" ]; then
    if [ ! -f ${CLUSTER_SECRETS_DIR}/gcs_keyfile.json ]; then
        echo "please ensure ${GCS_KEYFILE_PATH} exists"
    else
        cp ${GCS_KEYFILE_PATH} ${CLUSTER_SECRETS_DIR}
    fi
else
    echo "placeholder" > ${CLUSTER_SECRETS_DIR}/gcs_keyfile.json
fi

# create certificates
openssl genrsa -out ${CLUSTER_SECRETS_DIR}/private_key.pem 4096
openssl req -new -subj "/C=${ISO_3166_COUNTRY_CODE}/ST=${STATE}/L=${CITY}/O=DevOps/CN=${REGION}-${PROVIDER}.${SUBDOMAIN}" -x509 -key ${CLUSTER_SECRETS_DIR}/private_key.pem -out ${CLUSTER_SECRETS_DIR}/root.crt -days 3650

# create values.yaml
PRIVATE_KEY=$(cat ${CLUSTER_SECRETS_DIR}/private_key.pem | awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}')
ROOT_CERT=$(cat ${CLUSTER_SECRETS_DIR}/root.crt | awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}')

cat <<-EOF > ${CLUSTER_SECRETS_DIR}/values.yaml
    provider: $(echo ${PROVIDER})
    region: $(echo ${REGION})
    sharedKey: $(pwgen 16 1)
    adminserver:
        envVar:
            extEndpoint: https://$(echo ${EXTERNAL_REGISTRY_URL})
        emailPassword: $(pwgen 20 1)
        harborAdminPassword: $(pwgen 20 1)
    jobservice:
        secret: $(uuidgen)
    mysql:
        rootPassword: $(pwgen 20 1)
    postgres:
        rootPassword: $(pwgen 20 1)
    registry:
        auth:
            token:
                realm: https://$(echo ${EXTERNAL_REGISTRY_URL})/service/token
        httpSecret: $(pwgen 32 1)
        gcsKeyFile: $(cat ${CLUSTER_SECRETS_DIR}/gcs_keyfile.json)
    privateKey: "${PRIVATE_KEY}"
    rootCert: "${ROOT_CERT}"
    ui:
        secret: $(pwgen 32 1)
    nginx:
        annotations:
            external-dns.alpha.kubernetes.io/hostname: $(echo ${EXTERNAL_REGISTRY_URL}).
        sslCert: ${SSL_CERT}
        secretName: ${SUBDOMAIN}
EOF

# go get a cup of coffee
kubectl create namespace ${NAMESPACE}
kubectl -n common get secret ${SUBDOMAIN} -o json --export=true | \
kubectl -n ${NAMESPACE} create -f -
helm install . -f ${CLUSTER_SECRETS_DIR}/values.yaml --namespace ${NAMESPACE} --name harbor
