#!/bin/bash
set -x

set -e

function s3_to_https() {
  local s3_url="$1"

  if [[ "$s3_url" =~ ^s3://([^/]+)/(.+)$ ]]; then
    local bucket="${BASH_REMATCH[1]}"
    local path="${BASH_REMATCH[2]}"
    # current s3 bucket is create in this region
    local region="us-west-1"
    echo "https://${bucket}.s3.${region}.amazonaws.com/${path}"
  else
    echo "Invalid S3 URL: $s3_url" >&2
    return 1
  fi
}


function uploader {
    converted_url=$(s3_to_https "s3://$2/$1")
    echo "download url $converted_url"
    aws s3 cp $1 s3://$2/$1
}

function publishImage {
    echo "Publishing images to Docker Hub..."
    echo "The images on the host:"
    # for develop, will use 'dev' as the tag name
    # for release-*, will use 'release-*-dev' as the tag name, like release-v1.8.0-dev
    if [[ $1 == "develop" ]]; then
      image_tag=dev
    fi
    if [[ $1 == "release-"* ]]; then
      image_tag=$2-dev
    fi
    # if arch is provided (5th arg), append it as suffix for multi-arch support
    local arch_suffix=""
    if [ -n "${5:-}" ]; then
      arch_suffix="-$5"
    fi
    # rename the images with tag "dev" and push to Docker Hub
    docker images
    set +x
    printf '%s\n' "$4" | docker login --username "$3" --password-stdin
    set -x
    # format the output to be compatible with both old and new Docker versions.
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.CreatedAt}}\t{{.Size}}"| grep goharbor | grep -v "\-base" | sed -n "s|\(goharbor/[-._a-z0-9]*\)\s*\(.*$2\).*|docker tag \1:\2 \1:${image_tag}${arch_suffix};docker push \1:${image_tag}${arch_suffix}|p" | bash
    echo "Images are published successfully"
    docker images
}

function publishImageGhcr {
    local branch="$1"
    local version="$2"
    local ghcr_user="$3"
    local ghcr_token="$4"
    local arch="${5:-}"

    if [[ $branch == "develop" ]]; then
      image_tag=dev
    elif [[ $branch == "release-"* ]]; then
      image_tag="${version}-dev"
    fi

    local arch_suffix=""
    if [ -n "$arch" ]; then
      arch_suffix="-$arch"
    fi

    echo "Publishing images to ghcr.io/${ghcr_user}..."
    set +x
    printf '%s\n' "$ghcr_token" | docker login ghcr.io --username "$ghcr_user" --password-stdin
    set -x
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.CreatedAt}}\t{{.Size}}" \
      | grep goharbor | grep -v "\-base" | grep -v "harbor-db" | grep -v "valkey" | grep -v "photon" | grep -v "^goharbor/prepare" \
      | sed -n "s|\(goharbor/\([-._a-z0-9]*\)\)\s*\(.*${version}\).*|docker tag \1:${version} ghcr.io/${ghcr_user}/\2:${image_tag}${arch_suffix};docker push ghcr.io/${ghcr_user}/\2:${image_tag}${arch_suffix}|p" \
      | bash
    docker logout ghcr.io
    echo "Images published to ghcr.io/${ghcr_user} successfully"
}
