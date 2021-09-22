#!/bin/sh

set -e

/harbor/install_cert.sh

exec /harbor/harbor_jobservice -c /etc/jobservice/config.yml
