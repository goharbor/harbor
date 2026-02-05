#!/bin/sh

set -e

. /harbor/libredis.sh

/harbor/install_cert.sh

configure_redis_jobservice

exec /harbor/harbor_jobservice -c /etc/jobservice/config.yml
