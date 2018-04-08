#!/bin/sh
if [ -d /etc/ui/private ]; then
    chown -R 10000:10000 /etc/ui/private
fi
sudo -E -u \#10000 "/harbor/harbor_ui"

