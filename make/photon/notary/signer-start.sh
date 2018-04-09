#!/bin/sh
chown 10000:10000 -R /etc/notary/private
sudo -E -u \#10000 sh -c "/usr/bin/env && /migrations/migrate.sh && /bin/notary-signer -config=/etc/notary/signer-config.json -logf=logfmt"
