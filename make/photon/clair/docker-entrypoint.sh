#!/bin/bash
set -e
sudo -E -H -u \#10000 sh -c "/dumb-init -- /clair2.0.1/clair -config /etc/clair/config.yaml"
set +e
