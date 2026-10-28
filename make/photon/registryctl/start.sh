#!/bin/sh

set -e

# The directory /var/lib/registry is within the container, and used to store image in CI testing.
# So for now we need to chown to it to avoid failure in CI.
# if [ -d /var/lib/registry ]; then
#     chown 10000:10000 -R /var/lib/registry
# fi

# Quote REGISTRY_HTTP_SECRET to avoid yaml parsing error in distribution config parser
if [ -n "$REGISTRY_HTTP_SECRET" ]; then
    export REGISTRY_HTTP_SECRET="'"$(echo "$REGISTRY_HTTP_SECRET" | sed "s/'/''/g")"'"
fi

/home/harbor/install_cert.sh

exec /home/harbor/harbor_registryctl -c /etc/registryctl/config.yml
