#!/bin/sh
if [ -d /etc/adminserver ]; then
    chown -R 10000:10000 /etc/adminserver
fi
sudo -E -u \#10000 "/harbor/harbor_adminserver"
