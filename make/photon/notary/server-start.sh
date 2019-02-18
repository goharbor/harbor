#!/bin/sh
sudo -E -u \#10000 sh -c "/bin/notary-server -config=/etc/notary/server-config.postgres.json -logf=logfmt"
