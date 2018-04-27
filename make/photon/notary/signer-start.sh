#!/bin/sh
sudo -E -u \#10000 sh -c "/usr/bin/env && /migrations/migrate.sh && /bin/notary-signer -config=/etc/notary/signer-config.postgres.json -logf=logfmt"
