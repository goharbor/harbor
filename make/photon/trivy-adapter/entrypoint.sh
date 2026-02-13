#!/bin/sh

set -e

. /home/scanner/libredis.sh

/home/scanner/install_cert.sh

configure_redis_trivy

exec /home/scanner/bin/scanner-trivy
