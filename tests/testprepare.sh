#!/bin/bash
set -e
cp tests/docker-compose.test.yml make/.

mkdir -p /etc/ui
cp make/common/config/ui/private_key.pem /etc/ui/

mkdir src/ui/conf
cp make/common/config/ui/app.conf src/ui/conf/
if [ "$(uname)" == "Darwin" ]; then
	IP=`ifconfig en0 | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}'`
else 
        IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
fi
echo "server ip is "$IP
sed -i -r "s/POSTGRESQL_HOST=postgresql/POSTGRESQL_HOST=$IP/" make/common/config/adminserver/env
sed -i -r "s|REGISTRY_URL=http://registry:5000|REGISTRY_URL=http://$IP:5000|" make/common/config/adminserver/env
sed -i -r "s/UI_SECRET=.*/UI_SECRET=$UI_SECRET/" make/common/config/adminserver/env

chmod 777 /data/
