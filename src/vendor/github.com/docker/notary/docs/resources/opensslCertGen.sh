#!/usr/bin/env bash

# Script to be used for generating testing certs only - for a production system,
# a public CA or an internal CA should be used


CLIENT_USAGE=$(cat <<EOL
Generate a self-signed client cert and key to be used in mutual TLS.

${0} client [-o <output file prefix>]

Example:
    ${0} client -o clienttls
EOL
)

SERVER_USAGE=$(cat <<EOL
Generate a self-signed cert key and certificate.

${0} server -n <common name> [-o <output file prefix>]
    [-r <root key if do not want it self-signed>]
    [-a <subjectAltName>] [-a <subjectAltName>] ...

Example:
    ${0} server -o servertls -n notary-server -a DNS:notaryserver \
        -a DNS:notary_server -a IP:127.0.0.1
EOL
)

if [[ -z "${1}" ]]; then
    printf "${CLIENT_USAGE}\n\n${SERVER_USAGE}\n\n"
    exit 1
fi

OPENSSLCNF=
for path in /etc/openssl/openssl.cnf /etc/ssl/openssl.cnf /usr/local/etc/openssl/openssl.cnf; do
    if [[ -e ${path} ]]; then
        OPENSSLCNF=${path}
    fi
done
if [[ -z ${OPENSSLCNF} ]]; then
    printf "Could not find openssl.cnf"
    exit 1
fi

if [[ "${1}" == "client" ]]; then
    # Generate client keys - ensure that these keys are NOT CA's, otherwise
    # any client that is compromised can sign any number of other client
    # certs.
    OUT="clienttls"
    while getopts "o:" opt "${@:2}"; do
        case "${opt}" in
            o)
                OUT="${OPTARG}"
                ;;
            *)
                printf "${CLIENT_USAGE}"
                exit 1
                ;;
        esac
    done

    openssl genrsa -out "${OUT}.key" 4096
    openssl req -new -key "${OUT}.key" -out "${OUT}.csr" \
        -subj '/C=US/ST=CA/L=San Francisco/O=Docker/CN=Notary Testing Client Auth'

    cat > "${OUT}.cnf" <<EOL
[ssl_client]
basicConstraints = critical,CA:FALSE
nsCertType = critical, client
keyUsage = critical, digitalSignature, nonRepudiation
extendedKeyUsage = critical, clientAuth
authorityKeyIdentifier=keyid,issuer
EOL

    openssl x509 -req -days 3650 -in "${OUT}.csr" -signkey "${OUT}.key" \
        -out "${OUT}.crt" -extfile "${OUT}.cnf" -extensions ssl_client

    rm "${OUT}.cnf" "${OUT}.csr"
fi

if [[ "${1}" == "server" ]]; then
    # Create a server certificate

    OUT="servertls"
    COMMONNAME=
    SAN=
    while getopts ":o:n:a:" opt "${@:2}"; do
        case "${opt}" in
            o)
                OUT="${OPTARG}"
                ;;
            n)
                COMMONNAME="${OPTARG}"
                ;;
            a)
                SAN="${SAN} ${OPTARG}"
                ;;
            *)
                printf "${SERVER_USAGE}\n\n"
                exit 1
                ;;
        esac
    done

    if [[ -z "${COMMONNAME}" ]]; then
        printf "Please provide a common name/domain for the cert."
        printf "\n\n${SERVER_USAGE}\n\n"
        exit 1
    fi
    PPRINT_DOMAINS="${COMMONNAME}$(printf ", %s" "${SAN[@]}")"
    printf "Generating server certificate for domains: ${PPRINT_DOMAINS}\n\n"

    # see https://www.openssl.org/docs/manmaster/apps/x509v3_config.html for
    # more information on extensions
    cat "${OPENSSLCNF}" > "${OUT}.cnf"
    cat >> "${OUT}.cnf" <<EOL
[ v3_req ]
basicConstraints = critical,CA:FALSE
nsCertType = critical, server
keyUsage = critical, digitalSignature, nonRepudiation
extendedKeyUsage = critical, serverAuth
authorityKeyIdentifier=keyid,issuer
EOL
    if [[ -n "${SAN}" ]]; then
        printf "subjectAltName=$(echo ${SAN[@]} | tr ' ' ,)" >> "${OUT}.cnf"
    fi

    openssl genrsa -out "${OUT}.key" 4096
    openssl req -new -nodes -key "${OUT}.key" -out "${OUT}.csr" \
        -subj "/C=US/ST=CA/L=San Francisco/O=Docker/CN=${COMMONNAME}" \
        -config "${OUT}.cnf" -extensions "v3_req"
    openssl x509 -req -days 3650 -in "${OUT}.csr" -signkey "${OUT}.key" \
        -out "${OUT}.crt" -extensions v3_req -extfile "${OUT}.cnf"
fi
