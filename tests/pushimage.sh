#!/bin/bash
set -x
set -e

IP=$(hostname -I | awk '{print $1}')
docker pull registry.goharbor.io/dockerhub/library/hello-world:latest
docker pull registry.goharbor.io/dockerhub/library/busybox:latest
docker login -u admin -p Harbor12345 $IP:5000  

docker tag registry.goharbor.io/dockerhub/library/hello-world:latest $IP:5000/library/hello-world:latest
docker push $IP:5000/library/hello-world:latest
echo "$? pushed hello world"

docker tag registry.goharbor.io/dockerhub/library/busybox:latest $IP:5000/library/busybox:latest
docker push $IP:5000/library/busybox:latest
echo "$? pushed busybox"
