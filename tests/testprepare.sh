#!/bin/bash
set -e
cp tests/docker-compose.test.yml make/.

mkdir -p core
cp make/common/config/core/private_key.pem /etc/core/

mkdir src/core/conf
cp make/common/config/core/app.conf src/core/conf/
if [ "$(uname)" == "Darwin" ]; then
	IP=`ifconfig en0 | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}'`
else 
        IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
fi
echo "server ip is "$IP
chmod 777 /data/
