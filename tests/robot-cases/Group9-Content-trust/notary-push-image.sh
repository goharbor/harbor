#!/bin/bash

docker pull tomcat

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://$IP:4443

docker login -u admin -p Harbor12345 $IP

docker tag tomcat $IP/library/tomcat:latest
python ./tests/robot-cases/Group9-Content-trust/notary-push-image.py
