#!/bin/bash

echo "docker-compose version 1.7.1"
cd "$( dirname "${BASH_SOURCE[0]}" )" 
cp ./docker-compose-Linux-x86_64 /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

