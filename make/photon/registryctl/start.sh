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

/harbor/install_cert.sh

sudo -E -u \#10000 "/harbor/harbor_registryctl" "-c" "/etc/registryctl/config.yml"
