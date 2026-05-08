#!/bin/bash
set -x
set -e

profile_path=$(pwd)/profile.cov
echo "profile.cov path: $profile_path"

cd $(dirname $(find . -name go.mod))

mapfile -t packages < <(go list ./... | grep -v -E 'tests|testing')
echo "testing packages: ${packages[*]}"

# Build a single coverpkg list from all harbor-internal packages so that
# cross-package coverage is tracked, matching the previous per-package behavior.
coverpkg=$(IFS=','; echo "${packages[*]}")

# Run packages in parallel; -p controls how many test binaries execute
# concurrently.  The -race detector forces covermode=atomic automatically.
NPROC=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

go test -race -v \
    -covermode=atomic \
    -coverprofile="$profile_path" \
    -coverpkg="$coverpkg" \
    -p "$NPROC" \
    "${packages[@]}"
