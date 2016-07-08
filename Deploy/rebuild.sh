#!/bin/bash
docker-compose stop
docker ps -a|grep ago | awk '{print $1}' |xargs docker rm
docker rmi deploy_ui
docker-compose up -d
