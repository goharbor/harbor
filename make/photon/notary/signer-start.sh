#!/bin/sh
sudo -E -u \#10000 sh -c "migrate-patch -database='${DB_URL}' && /migrations/migrate.sh && /bin/notary-signer -config=/etc/notary/signer-config.postgres.json -logf=logfmt"
