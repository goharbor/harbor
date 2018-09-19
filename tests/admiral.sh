#!/bin/bash

# run admiral for unit test
name=admiral
port=8282
docker rm -f $name 2>/dev/null
docker run -d -p $port:8282 --name $name vmware/admiral:v1.2.1

# solution user token file for test
mkdir -p /etc/core/token/
echo "token" > /etc/core/token/tokens.properties
