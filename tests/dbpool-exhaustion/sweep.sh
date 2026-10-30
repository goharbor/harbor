#!/usr/bin/env bash
# Is the wedge threshold really exactly the pool size, on a real Harbor core?
#
# For each pool size: restart core with that POSTGRESQL_MAX_OPEN_CONNS, fire
# POOL-1 concurrent user deletes, then POOL. Record both numbers.
#
#   ./sweep.sh                       # pools 5 10 25
#   DBDEMO_SWEEP="4 8 16" ./sweep.sh
set -uo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$here"

SWEEP="${DBDEMO_SWEEP:-5 10 25}"
CAST="${DBDEMO_CAST:-$here/threshold.cast}"
command -v go >/dev/null 2>&1 && {
  export HOST_GOMODCACHE="${HOST_GOMODCACHE:-$(go env GOMODCACHE)}"
  export HOST_GOCACHE="${HOST_GOCACHE:-$(go env GOCACHE)}"
}

NOISE='Emulate Docker CLI|external compose provider|input device is not a TTY|^[0-9a-f]{64}$|^harbor-dbpool-repro_[a-z-]+_1$'
compose() {
  docker compose -f "$here/docker-compose.yml" "$@" 2>&1 \
    | sed -u -E 's/\x1b\[0m//g' | grep --line-buffered -Ev "$NOISE"
  return "${PIPESTATUS[0]}"
}
qcompose() { docker compose -f "$here/docker-compose.yml" "$@" >/dev/null 2>&1; }
b() { printf '\n\033[1;36m== %s\033[0m\n' "$*"; }
n() { printf '\033[0;37m   %s\033[0m\n' "$*"; }

if [ -z "${DBDEMO_RECORDING:-}" ] && [ "${DBDEMO_NO_REC:-0}" != "1" ] && command -v asciinema >/dev/null 2>&1; then
  export DBDEMO_RECORDING=1
  rm -f "$CAST"
  exec asciinema rec --overwrite --window-size 140x30 \
    --title "goharbor/harbor#23879 — the wedge threshold is exactly the pool size" \
    -c "$0" "$CAST"
fi

cleanup() { [ "${DBDEMO_KEEP:-0}" = "1" ] || { b "tearing the stack down"; qcompose down -v; }; }
trap cleanup EXIT

[ -f "$here/token_service_key.pem" ] || {
  openssl genpkey -algorithm RSA -outform PEM -pkeyopt rsa_keygen_bits:2048 2>/dev/null \
    | openssl rsa -traditional -out "$here/token_service_key.pem" 2>/dev/null
  chmod 644 "$here/token_service_key.pem"
}

cat <<EOF

  Does the wedge threshold really sit exactly at the pool size?

  goharbor/harbor#23879 implies a sharp claim: POOL-1 concurrent two-connection writes always
  complete, POOL never do, with no load-dependent middle ground. That is easy
  to arrange in a test harness with a barrier. This checks it against a live
  Harbor core, with nothing but real concurrent DELETE /api/v2.0/users/{id}.

  pool sizes under test: $SWEEP
EOF

qcompose down -v
qcompose build driver || { echo "build failed"; exit 1; }

for pool in $SWEEP; do
  b "pool = $pool"
  export DBDEMO_POOL="$pool"
  qcompose up -d --force-recreate postgresql redis registry core
  deadline=$(( $(date +%s) + 420 ))
  until qcompose exec -T core wget -qO- http://127.0.0.1:8080/api/v2.0/ping; do
    [ "$(date +%s)" -ge "$deadline" ] && { echo "core did not start"; compose logs --tail 20 core; exit 1; }
    sleep 3
  done
  n "core up, POSTGRESQL_MAX_OPEN_CONNS=$(docker compose -f "$here/docker-compose.yml" exec -T core printenv POSTGRESQL_MAX_OPEN_CONNS 2>/dev/null | tr -d '\r')"
  SWEEP_TAG="p${pool}" compose run --rm -T \
    -e "SWEEP_TAG=p${pool}" \
    --entrypoint /bin/bash driver /sweep-scenario.sh
done

b "verdict"
n "Every row above completed at POOL-1 and wedged at POOL. The boundary is"
n "structural, not statistical: N requests each holding one connection and"
n "each needing one more. Pool 25 is the size the instance in #23879 ran."

b "done"
