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

# Split packages into two groups based on whether their test files directly
# import the shared DB/test-harness packages.  DB-dependent tests must run
# serially (-p 1) because they all share a single PostgreSQL instance and
# running them concurrently causes data-pollution / count-mismatch failures.
# Pure (mock-based) packages have no such constraint and run in parallel.
db_markers="github.com/goharbor/harbor/src/common/dao|github.com/goharbor/harbor/src/common/utils/test"

mapfile -t db_pkgs < <(
    go list -f '{{.ImportPath}}|{{range .TestImports}}{{.}} {{end}}{{range .XTestImports}}{{.}} {{end}}' \
        "${packages[@]}" \
    | grep -E "$db_markers" \
    | cut -d'|' -f1
)

mapfile -t pure_pkgs < <(
    go list -f '{{.ImportPath}}|{{range .TestImports}}{{.}} {{end}}{{range .XTestImports}}{{.}} {{end}}' \
        "${packages[@]}" \
    | grep -vE "$db_markers" \
    | cut -d'|' -f1
)

echo "DB-dependent packages (serial, -p 1): ${db_pkgs[*]}"
echo "Pure packages (parallel, -p $NPROC): ${pure_pkgs[*]}"

run_args=(
    -race -v
    -covermode=atomic
    -coverpkg="$coverpkg"
)

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
    head -1 "${profile_path}.pure" 2>/dev/null || head -1 "${profile_path}.db"
    for f in "${profile_path}.pure" "${profile_path}.db"; do
        [ -f "$f" ] && tail -n +2 "$f"
    done
} > "$profile_path"
rm -f "${profile_path}.pure" "${profile_path}.db"
