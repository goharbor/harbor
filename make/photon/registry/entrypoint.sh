#!/bin/sh

set -e

# The directory /var/lib/registry is within the container, and used to store image in CI testing.
# So for now we need to chown to it to avoid failure in CI.
if [ -d /var/lib/registry ]; then
    chown 10000:10000 -R /var/lib/registry
fi

if [ -d /storage ]; then
    if ! stat -c '%u:%g' /storage | grep -q '10000:10000' ; then
        # 10000 is the id of harbor user/group.
        # Usually NFS Server does not allow changing owner of the export directory,
        # so need to skip this step and requires NFS Server admin to set its owner to 10000.
        chown 10000:10000 -R /storage
    fi
fi

if [ ! -f /etc/pki/tls/certs/ca-bundle.crt.original ]; then
    cp /etc/pki/tls/certs/ca-bundle.crt /etc/pki/tls/certs/ca-bundle.crt.original
fi

if [ -f /etc/registry/custom-ca-bundle.crt ]; then
    if grep -q "Photon" /etc/lsb-release; then
        echo "Appending custom ca bundle ..."
        cp /etc/pki/tls/certs/ca-bundle.crt.original /etc/pki/tls/certs/ca-bundle.crt
        cat /etc/registry/custom-ca-bundle.crt >> /etc/pki/tls/certs/ca-bundle.crt
        echo "Done."
    else
        echo "Current OS is not Photon, skip appending ca bundle"
    fi
fi

case "$1" in
    *.yaml|*.yml) set -- registry serve "$@" ;;
    serve|garbage-collect|help|-*) set -- registry "$@" ;;
esac

sudo -E -u \#10000 "$@"
