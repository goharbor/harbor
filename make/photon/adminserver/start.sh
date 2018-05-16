#!/bin/sh

#In the case when the config store is set to filesystem, the directory has to be writable.
if [ -d /etc/adminserver/config ]; then
    chown -R 10000:10000 /etc/adminserver/config
fi
sudo -E -u \#10000 "/harbor/harbor_adminserver"
