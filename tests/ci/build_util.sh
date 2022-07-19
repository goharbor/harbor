#!/bin/bash
set -x

set -e

function uploader {
    gsutil cp $1 gs://$2/$1
    gsutil -D setacl public-read gs://$2/$1 &> /dev/null
}

function publishImage {
    echo "Publishing images to Docker Hub..."
    echo "The images on the host:"
    # for main, will use 'dev' as the tag name
    # for release-*, will use 'release-*-dev' as the tag name, like release-v1.8.0-dev
    if [[ $1 == "main" ]]; then
      image_tag=dev
    fi
    if [[ $1 == "release-"* ]]; then
      image_tag=$2-dev
    fi
    # rename the images with tag "dev" and push to Docker Hub
    docker images
    docker login -u $3 -p $4
    docker images | grep goharbor | grep -v "\-base" | sed -n "s|\(goharbor/[-._a-z0-9]*\)\s*\(.*$2\).*|docker tag \1:\2 \1:$image_tag;docker push \1:$image_tag|p" | bash
    echo "Images are published successfully"
    docker images
}