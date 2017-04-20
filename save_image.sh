#!/bin/bash
set -e

images=(deploy_ui deploy_jobservice deploy_mysql deploy_log nginx:1.9 registry:2.5.0)

mkdir -p images && cd images

for image in ${images[@]}; do
    echo "saving the image of ${image}"
    docker save ${image} >  ${image}.tar 
    echo "finished saving the image of ${image}"
done
