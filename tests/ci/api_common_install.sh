#!/bin/bash
set -x

set +e
sudo rm -fr /data/*
sudo mkdir -p /data
DIR="$(cd "$(dirname "$0")" && pwd)"

set -e
if [ -z "$1" ]; then echo no ip specified; exit 1;fi
# prepare cert ...
sudo ./tests/generateCerts.sh $1

python --version
pip -V
cat /etc/issue
cat /proc/version
sudo -H pip install --ignore-installed urllib3 chardet requests --upgrade
python --version

sudo ./tests/hostcfg.sh

if [ "$2" = 'LDAP' ]; then
    cd tests && sudo ./ldapprepare.sh && cd ..
fi

if [ $GITHUB_TOKEN ];
then
    sed "s/# github_token: xxx/github_token: $GITHUB_TOKEN/" -i make/harbor.yml
fi

sed "s|# metric:|metric:|" -i make/harbor.yml
sed "s|#   enabled: false|  enabled: true|" -i make/harbor.yml
sed "s|#   port: 9090|  port: 9090|" -i make/harbor.yml
sed "s|#   path: /metrics|  path: /metrics|" -i make/harbor.yml

sudo make build_base_docker compile build prepare COMPILETAG=compile_golangimage GOBUILDTAGS="include_oss include_gcs" BUILDBIN=true NOTARYFLAG=true TRIVYFLAG=true CHARTFLAG=true GEN_TLS=true

# set the debugging env
echo "GC_TIME_WINDOW_HOURS=0" | sudo tee -a ./make/common/config/core/env
sudo make start

# waiting 5 minutes to start
for((i=1;i<=30;i++)); do
  echo $i waiting 10 seconds...
  sleep 10
  curl -k -L -f 127.0.0.1/api/v2.0/systeminfo && break
  docker ps
done
