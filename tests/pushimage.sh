#!/bin/bash
docker pull hello-world
docker pull docker
docker login -u admin -p Harbor12345 127.0.0.1:5000  

docker tag hello-world 127.0.0.1:5000/library/hello-world
docker push 127.0.0.1:5000/library/hello-world

docker tag docker 127.0.0.1:5000/library/docker
docker push 127.0.0.1:5000/library/docker
