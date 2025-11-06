#!/bin/bash
set -x
set -e

function s3_to_https() {
  local s3_url="$1"
  if [[ "$s3_url" =~ ^s3://([^/]+)/(.+)$ ]]; then
    local bucket="${BASH_REMATCH[1]}"
    local path="${BASH_REMATCH[2]}"
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
  aws s3 cp "$1" "s3://$2/$1"
}

# NEW: persist the repo list from this runner
# Writes one line per repo (no tags), e.g. goharbor/harbor-core
function saveRepoList() {
  local outfile="$1"
  docker images --format "{{.Repository}}" \
    | grep '^goharbor/' \
    | grep -v '\-base' \
    | sort -u > "$outfile"
  echo "Saved repo list to $outfile"
  cat "$outfile"
}

# UPDATED: arch-aware publish
# Usage: publishImage <branch> <version> <docker_user> <docker_pass> <arch>
#  - main      -  base_tag=dev
#  - release-* - base_tag=<version>-dev
# Pushes: <repo>:<base_tag>-<arch>   (e.g. core:dev-amd64 / core:dev-arm64)
function publishImage {
  branch=$1
  version=$2
  user=$3
  pass=$4
  arch=$5

  if [[ "$branch" == "main" ]]; then
    base_tag="dev"
  elif [[ "$branch" == release-* ]]; then
    base_tag="${version}-dev"
  else
    base_tag="${version}"
  fi

  arch_tag="${base_tag}-${arch}"
  echo "Publishing images for arch=${arch}; tag=${arch_tag}"

  docker login -u "$user" -p "$pass"

  # Retag & push every non-base goharbor image we built on this runner that already carries $version
  docker images --format '{{.Repository}}:{{.Tag}}' \
    | grep '^goharbor/' \
    | grep -v '\-base' \
    | awk -F: -v version="$version" -v tag="$arch_tag" '
        $2 ~ version {
          printf("docker tag %s:%s %s:%s\n", $1, $2, $1, tag);
          printf("docker push %s:%s\n", $1, tag);
        }' | bash

  docker logout
  echo "Done pushing arch tag: ${arch_tag}"
}
