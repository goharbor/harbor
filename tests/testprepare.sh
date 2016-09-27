#!/bin/bash

cp tests/docker-compose.test.yml Deploy/.

mkdir /etc/ui
cp Deploy/config/ui/private_key.pem /etc/ui/.

mkdir conf
cp Deploy/config/ui/app.conf conf/.
