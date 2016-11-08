#!/bin/bash

dir=harbor_logs
mkdir -p $dir

echo "Version" >> $dir/docker
docker version >> $dir/docker
printf "\n\nInfo\n" >> $dir/docker
docker info >> $dir/docker
printf "\n\nImages\n" >> $dir/docker
docker images >> $dir/docker
printf "\n\nRunning containers\n" >> $dir/docker
docker ps >> $dir/docker

docker-compose version >> $dir/docker-compose

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cp -r $base_dir/../harbor/common $dir/
cp $base_dir/../harbor/harbor.cfg $dir/
cp -r /var/log/harbor $dir/
tar --remove-files -zcf $dir.tar.gz $dir