#!/usr/bin/env bash
# Stand up a real Harbor core with a small DB pool, wedge it the way
# goharbor/harbor#23879 was wedged in production, and record the whole thing.
#
#   ./run.sh                   # pool of 5
#   DBDEMO_POOL=8 ./run.sh
#   DBDEMO_NO_REC=1 ./run.sh   # no asciinema, just run it
#
# Everything runs in containers under the compose project `harbor-dbpool-repro`. No host
# ports are published, so this cannot collide with `task dev:up` or another rig.
set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$here"

export DBDEMO_POOL="${DBDEMO_POOL:-5}"
# Hand the container the host Go caches if this machine has them; without
# them compose falls back to named volumes and just downloads the modules.
command -v go >/dev/null 2>&1 && {
  export HOST_GOMODCACHE="${HOST_GOMODCACHE:-$(go env GOMODCACHE)}"
  export HOST_GOCACHE="${HOST_GOCACHE:-$(go env GOCACHE)}"
}
CAST="${DBDEMO_CAST:-$here/dbpool-repro.cast}"

# Podman's docker shim prints two banner lines on every invocation; they say
# nothing about the demo, so keep them out of the recording.
NOISE='Emulate Docker CLI|external compose provider|input device is not a TTY|^[0-9a-f]{64}$|^harbor-dbpool-repro_[a-z-]+_1$|^\\x1b\\[0m[0-9a-f]{64}$'
compose() {
  docker compose -f "$here/docker-compose.yml" "$@" 2>&1 \
    | sed -u -E 's/\x1b\[0m//g' \
    | grep --line-buffered -Ev "$NOISE"
  return "${PIPESTATUS[0]}"
}
qcompose() { docker compose -f "$here/docker-compose.yml" "$@" >/dev/null 2>&1; }

b() { printf '\n\033[1;36m== %s\033[0m\n' "$*"; }
n() { printf '\033[0;37m   %s\033[0m\n' "$*"; }

# Re-exec under asciinema unless we already are, or recording was declined.
if [ -z "${DBDEMO_RECORDING:-}" ] && [ "${DBDEMO_NO_REC:-0}" != "1" ]; then
  if command -v asciinema >/dev/null 2>&1; then
    export DBDEMO_RECORDING=1
    rm -f "$CAST"
    exec asciinema rec --overwrite --window-size 118x40 \
      --title "goharbor/harbor#23879 — DB pool deadlock on a real Harbor core (pool=$DBDEMO_POOL)" \
      -c "$0" "$CAST"
  fi
  echo "asciinema not found; running without a recording." >&2
fi

cleanup() {
  if [ "${DBDEMO_KEEP:-0}" = "1" ]; then
    b "leaving the stack up (DBDEMO_KEEP=1)"
    n "tear down with: docker compose -f $here/docker-compose.yml down -v"
    return
  fi
  b "tearing the stack down"
  qcompose down -v
}
trap cleanup EXIT

b "0. token signing key"
if [ ! -f "$here/token_service_key.pem" ]; then
  openssl genpkey -algorithm RSA -outform PEM -pkeyopt rsa_keygen_bits:4096 2>/dev/null \
    | openssl rsa -traditional -out "$here/token_service_key.pem" 2>/dev/null
  chmod 644 "$here/token_service_key.pem"
fi
n "$here/token_service_key.pem"

b "1. bringing up PostgreSQL, Redis, registry and a real Harbor core"
n "core is built from this checkout; its pool is POSTGRESQL_MAX_OPEN_CONNS=$DBDEMO_POOL"
qcompose down -v
qcompose build || { echo "image build failed"; compose build; exit 1; }
qcompose up -d postgresql redis registry core || { compose up -d postgresql redis registry core; exit 1; }

b "2. waiting for core to compile and serve"
t0=$(date +%s)
deadline=$(( $(date +%s) + 420 ))
until qcompose exec -T core wget -qO- http://127.0.0.1:8080/api/v2.0/ping >/dev/null 2>&1; do
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "core did not come up in time; last 40 log lines:"
    compose logs --tail 40 core
    exit 1
  fi
  sleep 3
done
n "core came up in $(( $(date +%s) - t0 ))s (compile + migrate + listen)"
n "core is serving:  $(docker compose -f "$here/docker-compose.yml" exec -T core wget -qO- http://127.0.0.1:8080/api/v2.0/ping 2>/dev/null)"
n "pool as core sees it: POSTGRESQL_MAX_OPEN_CONNS=$(docker compose -f "$here/docker-compose.yml" exec -T core printenv POSTGRESQL_MAX_OPEN_CONNS 2>/dev/null | tr -d '\r')"

b "3. the scenario"
compose run --rm -T driver
rc=$?

b "4. what core logged as it went under"
n "Filtered to the DELETEs and the audit-log config read, timestamps stripped."
n "Every wedged request gets as far as log.Middleware's audit PreCheck — the"
n "config lookup that sits immediately before userIDToName reaches for the"
n "second connection — and stops there:"
docker compose -f "$here/docker-compose.yml" logs core 2>&1 \
  | sed -E 's/\x1b\[[0-9;]*m//g' \
  | grep -E 'DELETE /api/v2.0/users/[0-9]+|disabled_audit_log_event_types' \
  | sed -E 's/^.*\| *//; s/^[0-9T:Z.-]+ //; s/\[DEBUG\] //; s/\[[^]]*middleware[^]]*\]: //' \
  | sed -E 's/attach request id [0-9a-f-]+ to the logger for the request //' \
  | uniq -c | tail -12 | sed 's/^/   /'
n ""
n "Then nothing. No 500, no timeout, no rollback, no error of any kind — the"
n "requests simply never come back. An operator watching core logs sees a"
n "healthy process that has gone quiet."

b "5. recovery — only a restart clears it"
n "No client, no request deadline and no pool timeout can release those five"
n "transactions. Restarting core can, because killing the process is what drops"
n "the connections."
qcompose restart core
deadline=$(( $(date +%s) + 300 ))
until qcompose exec -T core wget -qO- http://127.0.0.1:8080/api/v2.0/ping; do
  [ "$(date +%s)" -ge "$deadline" ] && { echo "core did not restart in time"; break; }
  sleep 3
done
compose run --rm -T --entrypoint /bin/bash driver -c '
  set -u
  for ep in health projects systeminfo; do
    printf "   GET /api/v2.0/%-11s -> %s\n" "$ep" \
      "$(curl -s -m 10 -o /dev/null -w "HTTP %{http_code} in %{time_total}s" -u admin:Harbor12345 "$CORE/api/v2.0/$ep")"
  done
  echo
  echo "   sessions idle in transaction after the restart:"
  psql -X -c "SELECT count(*) AS idle_in_transaction FROM pg_stat_activity
               WHERE datname = current_database() AND state = '"'"'idle in transaction'"'"'
                 AND pid <> pg_backend_pid();"
  echo "   users the wedged DELETEs were supposed to remove:"
  curl -s -u admin:Harbor12345 "$CORE/api/v2.0/users?page_size=100" \
    | jq -r "   [.[] | select(.username | startswith(\"victim\")) | .username] | \"   \(length) of them are still there: \(join(\", \"))\""'

b "done"
exit $rc
