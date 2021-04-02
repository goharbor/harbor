#!/bin/sh

if [ $(cat /$HOME/.docker/config.json |grep experimental |wc -l) -eq 0 ];then
    sed -i '$d' /$HOME/.docker/config.json
    sed -i '$d' /$HOME/.docker/config.json
    echo -e "},\n        \"experimental\": \"enabled\"\n}" >> /$HOME/.docker/config.json
fi
