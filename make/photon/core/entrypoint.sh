#!/bin/sh

set -e

. /harbor/libredis.sh

/harbor/install_cert.sh

configure_redis_core

exec /harbor/harbor_core
