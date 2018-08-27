#!/bin/bash
set -e

#/chart_storage is the directory in the contaienr for storing the chart artifacts
#if storage driver is set to 'local'
if [ -d /chart_storage ]; then
    chown 10000:10000 -R /chart_storage
fi

#Config the custom ca bundle
if [ -f /etc/chartserver/custom-ca-bundle.crt ]; then
    if grep -q "Photon" /etc/lsb-release; then
        if [ ! -f /etc/pki/tls/certs/ca-bundle.crt.original ]; then
            cp /etc/pki/tls/certs/ca-bundle.crt /etc/pki/tls/certs/ca-bundle.crt.original
        fi

        echo "Appending custom ca bundle ..."
        cp /etc/pki/tls/certs/ca-bundle.crt.original /etc/pki/tls/certs/ca-bundle.crt
        cat /etc/chartserver/custom-ca-bundle.crt >> /etc/pki/tls/certs/ca-bundle.crt
        echo "Done."
    else
        echo "Current OS is not Photon, skip appending ca bundle"
    fi
fi

#Start the server process
sudo -E -H -u \#10000 sh -c "/chartserver/chartm" #Parameters are set by ENV
set +e
