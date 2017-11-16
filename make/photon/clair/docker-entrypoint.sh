#!/bin/bash
set -e
/dumb-init -- /clair2.0.1/clair -config /config/config.yaml
set +e
