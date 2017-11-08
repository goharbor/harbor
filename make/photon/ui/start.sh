#!/bin/sh
if [ -d /etc/ui/ ]; then
    chown -R 10000:10000 /etc/ui/ 
fi
sudo -E -u \#10000 "/harbor/harbor_ui"

