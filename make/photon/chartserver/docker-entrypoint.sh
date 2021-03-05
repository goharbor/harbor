#!/bin/bash
set -e


/home/chart/install_cert.sh

#Start the server process
exec /home/chart/chartm
