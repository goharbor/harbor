#!/bin/bash
set -e
chown -R 10000:10000 /config
sudo -E -H -u \#10000 sh -c "/dumb-init -- /clair2.0.1/clair -config /config/config.yaml"
set +e
