#!/bin/bash
set -e

#/chart_storage is the directory in the contaienr for storing the chart artifacts
#if storage driver is set to 'local'
if [ -d /chart_storage ]; then
    chown 10000:10000 -R /chart_storage
fi

/harbor/install_cert.sh

#Start the server process
sudo -E -H -u \#10000 sh -c "/chartserver/chartm" #Parameters are set by ENV
set +e
