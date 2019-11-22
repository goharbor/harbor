#!/bin/bash
set -e


/home/chart/install_cert.sh

#Start the server process
/home/chart/chartm

set +e
