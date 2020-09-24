#!/bin/sh

set -e

# The directory /var/lib/registry is within the container, and used to store image in CI testing.
# So for now we need to chown to it to avoid failure in CI.
# if [ -d /var/lib/registry ]; then
#     chown 10000:10000 -R /var/lib/registry
# fi

/home/harbor/install_cert.sh

/home/harbor/harbor_registryctl -c /etc/registryctl/config.yml
