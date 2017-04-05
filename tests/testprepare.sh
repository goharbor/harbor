#!/bin/bash
set -e
cp tests/docker-compose.test.yml make/.

mkdir -p /etc/ui
cp make/common/config/ui/private_key.pem /etc/ui/.

mkdir conf
cp make/common/config/ui/app.conf conf/.

sed -i -r "s/MYSQL_HOST=mysql/MYSQL_HOST=127.0.0.1/" make/common/config/adminserver/env
sed -i -r "s|REGISTRY_URL=http://registry:5000|REGISTRY_URL=http://127.0.0.1:5000|" make/common/config/adminserver/env
sed -i -r "s/UI_SECRET=.*/UI_SECRET=$UI_SECRET/" make/common/config/adminserver/env

chmod 777 /data/