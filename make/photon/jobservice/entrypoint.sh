#!/bin/sh

set -e

/harbor/install_cert.sh

/harbor/harbor_jobservice -c /etc/jobservice/config.yml
