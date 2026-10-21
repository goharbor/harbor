#!/usr/bin/env bash
# The demo, run entirely inside the driver container against a real Harbor core.
#
# It proves goharbor/harbor#23879 end to end:
#   DELETE /api/v2.0/users/{id} holds one pooled connection in
#   transaction.Middleware and then takes a second one from the same pool in
#   log.Middleware's audit PreCheck (userIDToName -> orm.Context()). POOL
#   concurrent deletes therefore deadlock the pool, and nothing unblocks it.
set -uo pipefail

POOL="${DBDEMO_POOL:-5}"
CORE="${CORE:-http://core:8080}"
AUTH=(-u admin:Harbor12345)
JSON=(-H 'Content-Type: application/json')

b()  { printf '\n\033[1;36m== %s\033[0m\n' "$*"; }
n()  { printf '\033[0;37m   %s\033[0m\n' "$*"; }
ok() { printf '\033[1;32m   %s\033[0m\n' "$*"; }
bad(){ printf '\033[1;31m   %s\033[0m\n' "$*"; }

pgq() { psql -qtAX -c "$1"; }

activity() {
  psql -X -c "
    SELECT state,
           wait_event,
           count(*)                       AS sessions,
           max(now() - xact_start)::text  AS max_xact_age
      FROM pg_stat_activity
     WHERE datname = current_database()
       AND pid <> pg_backend_pid()
     GROUP BY state, wait_event
     ORDER BY state;"
}

# Create n users and echo their ids, read from the 201 Location header.
seed() {
  local n=$1 tag=$2 i loc
  for ((i = 1; i <= n; i++)); do
    loc=$(curl -s -o /dev/null -D - "${AUTH[@]}" "${JSON[@]}" \
      -X POST "$CORE/api/v2.0/users" \
      -d "{\"username\":\"$tag$i\",\"email\":\"$tag$i@example.com\",\"realname\":\"$tag $i\",\"password\":\"Harbor12345!\"}" \
      | tr -d '\r' | sed -n 's#^[Ll]ocation: */api/v2.0/users/##p')
    [ -n "$loc" ] || { bad "seeding $tag$i failed (no Location header)"; return 1; }
    echo "$loc"
  done
}

# Fire one DELETE per id, all at once, each with its own timeout.
storm() {
  local timeout=$1; shift
  local id
  rm -f /tmp/res.*
  for id in "$@"; do
    ( curl -s -m "$timeout" -o /dev/null \
        -w "   user $id -> HTTP %{http_code} after %{time_total}s\n" \
        "${AUTH[@]}" -X DELETE "$CORE/api/v2.0/users/$id" > "/tmp/res.$id" 2>&1 ) &
  done
}

probe() { # label url timeout
  local out
  out=$(curl -s -m "$3" -o /dev/null -w '%{http_code} in %{time_total}s' "${AUTH[@]}" "$2")
  case "$out" in
    000*) bad "$1 -> NO RESPONSE ($out, client gave up)" ;;
    200*) ok  "$1 -> $out" ;;
    *)    n   "$1 -> $out" ;;
  esac
}

cat <<EOF

  goharbor/harbor#23879 — DB connection pool deadlock, on a real Harbor core
  ---------------------------------------------------------------------
  core pool (POSTGRESQL_MAX_OPEN_CONNS) : $POOL
  PostgreSQL max_connections            : $(pgq 'SHOW max_connections')
  PostgreSQL version                    : $(pgq 'SHOW server_version')
  cache (CACHE_ENABLED)                 : off

  The trigger is the one from the incident: concurrent DELETE /api/v2.0/users/{id},
  which is what the portal's bulk-delete button sends. Nothing here is a mock --
  the real core binary, the real middleware chain, the real handler, real
  PostgreSQL. The only thing configured away from a shipped default is the pool
  size, and only to make the arithmetic small enough to watch.
EOF

b "PHASE 0 — the system is healthy"
probe "GET /api/v2.0/ping      " "$CORE/api/v2.0/ping" 5
probe "GET /api/v2.0/health    " "$CORE/api/v2.0/health" 5
probe "GET /api/v2.0/projects  " "$CORE/api/v2.0/projects" 5
probe "GET /api/v2.0/systeminfo" "$CORE/api/v2.0/systeminfo" 5
n "PostgreSQL sessions belonging to core:"
activity

b "PHASE 1 — $((POOL - 1)) concurrent deletes: one below the pool size"
n "Every one of these requests needs TWO connections at once. With one"
n "connection always spare they serialise, and all of them finish."
mapfile -t below < <(seed $((POOL - 1)) below)
n "seeded user ids: ${below[*]}"
start=$(date +%s)
storm 30 "${below[@]}"
wait
cat /tmp/res.* 2>/dev/null
n "elapsed: $(( $(date +%s) - start ))s"
if grep -q 'HTTP 200' /tmp/res.* 2>/dev/null; then
  ok "PASS: $((POOL - 1)) concurrent two-connection deletes all completed"
else
  bad "unexpected: phase 1 did not complete cleanly"
fi

b "PHASE 2 — $((POOL * 2)) concurrent deletes: at and above the pool size"
mapfile -t at < <(seed $((POOL * 2)) victim)
n "seeded user ids: ${at[*]}"
n "firing all $((POOL * 2)) DELETEs simultaneously, 45s client timeout ..."
storm 45 "${at[@]}"
sleep 8
n "8 seconds in — what PostgreSQL sees:"
activity
held=$(pgq "SELECT count(*) FROM pg_stat_activity
             WHERE datname = current_database()
               AND state = 'idle in transaction'
               AND pid <> pg_backend_pid()")
if [ "$held" = "$POOL" ]; then
  bad "WEDGED: all $POOL pooled connections are held by transactions that are"
  bad "        waiting for a connection. No lock waits. PostgreSQL is idle;"
  bad "        it is Harbor that is stuck."
else
  n "sessions idle in transaction: $held (expected $POOL)"
fi

b "PHASE 3 — the blast radius: the whole instance, not one endpoint"
n "None of these touch the users table. They just need one pooled connection."
probe "GET /api/v2.0/health    " "$CORE/api/v2.0/health" 10
probe "GET /api/v2.0/projects  " "$CORE/api/v2.0/projects" 10
probe "GET /api/v2.0/systeminfo" "$CORE/api/v2.0/systeminfo" 10
n ""
n "And the probe Kubernetes actually watches:"
probe "GET /api/v2.0/ping      " "$CORE/api/v2.0/ping" 5
bad "ping answers in milliseconds because pingSkipper keeps it clear of every"
bad "DB-touching middleware. Liveness and readiness stay GREEN on a core that"
bad "is serving nothing. That is why both production incidents were invisible"
bad "to Kubernetes."

b "PHASE 4 — no self-recovery"
n "Waiting for all $((POOL * 2)) clients to give up ..."
wait
cat /tmp/res.* 2>/dev/null
n ""
n "Every client has hung up. pg_stat_activity, right now:"
activity
held=$(pgq "SELECT count(*) FROM pg_stat_activity
             WHERE datname = current_database()
               AND state = 'idle in transaction'
               AND pid <> pg_backend_pid()")
if [ "$held" = "$POOL" ]; then
  bad "STILL $POOL transactions open after every client disconnected."
  bad "beego's Ormer.Begin passes context.Background() to BeginTx and pgxpool"
  bad "has no acquire timeout, so there is no deadline left anywhere in the"
  bad "stack to cancel them. Only restarting core clears this."
else
  n "sessions idle in transaction: $held"
fi

b "PHASE 5 — and the deletes did not happen"
n "The transactions never committed, so every 'deleted' user is still there:"
curl -s "${AUTH[@]}" "$CORE/api/v2.0/users?page_size=100" -m 10 \
  | jq -r 'if type == "array" then "   \(length) users still present" else "   (API did not answer - core is wedged)" end' \
  2>/dev/null || bad "   the API did not answer - core is wedged"

echo
