#!/bin/bash
set -x

set -e

function uploader {
    gsutil cp $1 gs://$2/$1
    gsutil -D setacl public-read gs://$2/$1 &> /dev/null
}