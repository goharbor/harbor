#!/usr/bin/env bash

set -eu

[[ -z "${DEBUG:-}" ]] || set -x

main() {
  local dir
  dir="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

  "${dir}/bootstrap.sh"
  # shellcheck disable=SC1090
  source "$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)/.env"

  export APITEST_DB=true

  cd /home/travis/go/src/github.com/goharbor/harbor
  bash ./tests/travis/api_common_install.sh "$IP" DB
  bash ./tests/travis/api_run.sh DB "$IP"
}

main
