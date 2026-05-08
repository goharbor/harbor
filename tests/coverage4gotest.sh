#!/bin/bash
set -x
set -e

profile_path=$(pwd)/profile.cov
echo "profile.cov path: $profile_path"

echo "mode: atomic" > "$profile_path"

cd $(dirname $(find . -name go.mod))

# Collect all testable packages once, reusing the result for both
# classification and the actual test run.
mapfile -t packages < <(go list ./... | grep -v -E 'tests|testing')
echo "testing packages: ${packages[*]}"

# Build coverpkg: all harbor-internal packages, matching the scope used by
# the original per-package listDeps approach but computed in one shot.
coverpkg=$(IFS=','; echo "${packages[*]}")

# Run all packages serially (-p 1) so that tests sharing a single
# PostgreSQL / Redis instance do not interfere with each other.
# A single `go test ./...` invocation allows the Go toolchain to reuse
# compilation artifacts across packages, which is faster than the old
# per-package loop even though execution is still sequential.
go test -race -v \
    -covermode=atomic \
    -coverprofile="${profile_path}.tmp" \
    -coverpkg="$coverpkg" \
    -p 1 \
    "${packages[@]}"

# Append coverage data (skip the mode header already written above).
tail -n +2 "${profile_path}.tmp" >> "$profile_path"
rm -f "${profile_path}.tmp"
