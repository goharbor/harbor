#!/bin/bash
set -e

for image in `ls images`; do
    echo "loading the image of ${image}"
    docker load < images/${image}
    echo "finished loading the image of ${image}"
done
