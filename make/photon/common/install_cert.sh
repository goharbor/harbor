#!/bin/sh

set -e

if [ ! -f ~/ca-bundle.crt.original ]; then
    cp /etc/pki/tls/certs/ca-bundle.crt ~/ca-bundle.crt.original
fi

cp ~/ca-bundle.crt.original /etc/pki/tls/certs/ca-bundle.crt

if [ "$(ls -A /harbor_cust_cert)" ]; then
    if grep -q "Photon" /etc/lsb-release; then
        echo "Appending trust CA to ca-bundle ..."
        for z in /harbor_cust_cert/*; do
            case ${z} in
                *.crt | *.ca | *.ca-bundle | *.pem)
                    if [ -d "$z" ]; then
                        echo "$z is dirictory, skip it ..."
                    else
                        cat $z >> /etc/pki/tls/certs/ca-bundle.crt
                        echo " $z Appended ..."
                    fi
                    ;;
                *) echo "$z is Not ca file ..." ;;
            esac
        done
        echo "CA appending is Done."
    else
        echo "Current OS is not Photon, skip appending ca bundle"
    fi
fi
