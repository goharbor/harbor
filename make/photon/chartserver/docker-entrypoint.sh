#!/bin/bash
set -e

#/chart_storage is the directory in the contaienr for storing the chart artifacts
#if storage driver is set to 'local'
if [ -d /chart_storage ]; then
    if ! stat -c '%u:%g' /chart_storage | grep -q '10000:10000' ; then
        # 10000 is the id of harbor user/group.
        # Usually NFS Server does not allow changing owner of the export directory,
        # so need to skip this step and requires NFS Server admin to set its owner to 10000.
        chown 10000:10000 -R /chart_storage
    fi
fi

/harbor/install_cert.sh

#Start the server process
sudo -E -H -u \#10000 sh -c "/chartserver/chartm" #Parameters are set by ENV
set +e
