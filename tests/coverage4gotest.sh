#!/bin/bash
set -x
set -e

profile_path=$(pwd)/profile.cov
echo "profile.cov path: $profile_path"

cd $(dirname $(find . -name go.mod))

mapfile -t packages < <(go list ./... | grep -v -E 'tests|testing')
echo "testing packages: ${packages[*]}"

# Build a single coverpkg list from all harbor-internal packages so that
# cross-package coverage is tracked across the whole codebase.
coverpkg=$(IFS=','; echo "${packages[*]}")

NPROC=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

# Classify packages into DB-dependent vs pure by grepping test source files
# directly.  This avoids an extra `go list` invocation and is not affected
# by `set -e` / pipefail behaviour.
#
# DB-dependent packages share a single PostgreSQL instance and must run
# serially (-p 1); running them concurrently causes data-pollution failures
# (count mismatches, schema-migration races).
# Pure (mock-based) packages have no such constraint and run in parallel.
declare -A db_dir_set
while IFS= read -r f; do
    db_dir_set["$(dirname "$f")"]=1
done < <(grep -rl \
    -E '"github\.com/goharbor/harbor/src/common/dao"|"github\.com/goharbor/harbor/src/common/utils/test"' \
    --include='*_test.go' . 2>/dev/null)

db_pkgs=()
pure_pkgs=()
for pkg in "${packages[@]}"; do
    rel="${pkg#github.com/goharbor/harbor/src/}"
    if [[ -n "${db_dir_set["./$rel"]+x}" ]]; then
        db_pkgs+=("$pkg")
    else
        pure_pkgs+=("$pkg")
    fi
done

echo "DB-dependent packages (serial, -p 1): ${#db_pkgs[@]}"
echo "Pure packages (parallel, -p $NPROC): ${#pure_pkgs[@]}"

run_args=(-race -v -covermode=atomic -coverpkg="$coverpkg")

# Run pure (non-DB) packages in parallel.
if [ ${#pure_pkgs[@]} -gt 0 ]; then
    go test "${run_args[@]}" \
        -coverprofile="${profile_path}.pure" \
        -p "$NPROC" \
        "${pure_pkgs[@]}"
fi

# Run DB-dependent packages serially to avoid shared-state conflicts.
if [ ${#db_pkgs[@]} -gt 0 ]; then
    go test "${run_args[@]}" \
        -coverprofile="${profile_path}.db" \
        -p 1 \
        "${db_pkgs[@]}"
fi

# Merge the two coverage profiles into the single file expected by Codecov.
{
    head -1 "${profile_path}.pure" 2>/dev/null \
        || head -1 "${profile_path}.db" 2>/dev/null
    for f in "${profile_path}.pure" "${profile_path}.db"; do
        [ -f "$f" ] && tail -n +2 "$f"
    done
} > "$profile_path"
rm -f "${profile_path}.pure" "${profile_path}.db"
