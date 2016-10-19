#!/bin/bash

cp tests/docker-compose.test.yml make/.

mkdir /etc/ui
cp make/common/config/ui/private_key.pem /etc/ui/.

mkdir conf
cp make/common/config/ui/app.conf conf/.
