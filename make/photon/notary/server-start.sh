#!/bin/sh
sudo -E -u \#10000 sh -c "/usr/bin/env /migrations/migrate.sh && /bin/notary-server -config=/etc/notary/server-config.postgres.json -logf=logfmt"
