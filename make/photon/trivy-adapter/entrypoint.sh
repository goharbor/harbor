#!/bin/sh

set -e

/home/scanner/install_cert.sh

exec /home/scanner/bin/scanner-trivy
