#!/bin/sh

if [ $(cat /$HOME/.docker/config.json |grep experimental |wc -l) -eq 0 ];then
    sudo sed -i '$d' /$HOME/.docker/config.json
    sudo sed -i '$d' /$HOME/.docker/config.json
    sudo echo -e "},\n        \"experimental\": \"enabled\"\n}" >> /$HOME/.docker/config.json
fi
