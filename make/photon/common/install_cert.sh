#!/bin/sh

set -e

if [ ! -f /etc/pki/tls/certs/ca-bundle.crt.original ]; then
    cp /etc/pki/tls/certs/ca-bundle.crt /etc/pki/tls/certs/ca-bundle.crt.original
fi

if [ -f /harbor_cust_cert/custom-ca-bundle.crt ]; then
    if grep -q "Photon" /etc/lsb-release; then
        echo "Appending custom ca bundle ..."
        cp /etc/pki/tls/certs/ca-bundle.crt.original /etc/pki/tls/certs/ca-bundle.crt
        cat /harbor_cust_cert/custom-ca-bundle.crt >> /etc/pki/tls/certs/ca-bundle.crt
        echo "Done."
    else
        echo "Current OS is not Photon, skip appending ca bundle"
    fi
fi