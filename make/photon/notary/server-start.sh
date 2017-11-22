#!/bin/sh
chown 10000:10000 -R /config
sudo -E -u \#10000 sh -c "/usr/bin/env /migrations/migrate.sh && /bin/notary-server -config=/config/server-config.json -logf=logfmt"
