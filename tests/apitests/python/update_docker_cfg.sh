#!/bin/sh

sudo sed -i '$d' /$HOME/.docker/config.json
sudo sed -i '$d' /$HOME/.docker/config.json
sudo echo -e "\n        },\n        \"experimental\": \"enabled\"\n}" >> /$HOME/.docker/config.json
