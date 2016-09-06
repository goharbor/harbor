#!/bin/bash

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
docker pull hello-world
docker pull docker
docker login -u admin -p Harbor12345 $IP:5000  

docker tag hello-world $IP:5000/library/hello-world:latest
docker push $IP:5000/library/hello-world:latest

docker tag docker $IP:5000/library/docker:latest
docker push $IP:5000/library/docker:latest
