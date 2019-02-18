#!/bin/sh
sudo -E -u \#10000 sh -c "/bin/notary-signer -config=/etc/notary/signer-config.postgres.json -logf=logfmt"
