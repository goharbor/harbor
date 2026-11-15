#!/usr/bin/env bash
# One threshold measurement against a live core whose pool is DBDEMO_POOL.
#
# Fires exactly POOL-1 concurrent DELETE /api/v2.0/users/{id}, then exactly
# POOL. PR #956 claims the boundary between those two is absolute. This checks
# that claim against a real Harbor rather than a test harness.
set -uo pipefail

POOL="${DBDEMO_POOL:-5}"
CORE="${CORE:-http://core:8080}"
TAG="${SWEEP_TAG:-s}"
AUTH=(-u admin:Harbor12345)

pgq() { psql -qtAX -c "$1"; }

# Create n users and echo their ids, read from the 201 Location header, so the
# sweep never has to page a users list that grows past the API's page cap.
seed() { # n tag -> ids
  local n=$1 tag=$2 i loc
  for ((i = 1; i <= n; i++)); do
    loc=$(curl -s -o /dev/null -D - -w '' "${AUTH[@]}" -H 'Content-Type: application/json' \
      -X POST "$CORE/api/v2.0/users" \
      -d "{\"username\":\"$tag$i\",\"email\":\"$tag$i@example.com\",\"realname\":\"$tag $i\",\"password\":\"Harbor12345!\"}" \
      | tr -d '\r' | sed -n 's#^[Ll]ocation: */api/v2.0/users/##p')
    [ -n "$loc" ] || { echo "seed $tag$i: no Location header" >&2; return 1; }
    echo "$loc"
  done
}

run_round() { # concurrency tag timeout -> "completed idle_in_tx"
  local n=$1 tag=$2 timeout=$3 id done=0
  mapfile -t ids < <(seed "$n" "$tag")
  [ "${#ids[@]}" -eq "$n" ] || { echo "seeded ${#ids[@]}, wanted $n" >&2; return 1; }
  rm -f /tmp/sw.*
  for id in "${ids[@]}"; do
    ( curl -s -m "$timeout" -o /dev/null -w '%{http_code}' "${AUTH[@]}" \
        -X DELETE "$CORE/api/v2.0/users/$id" > "/tmp/sw.$id" 2>&1 ) &
  done
  sleep "$((timeout - 2))"
  local held
  held=$(pgq "SELECT count(*) FROM pg_stat_activity
               WHERE datname = current_database() AND state = 'idle in transaction'
                 AND pid <> pg_backend_pid()")
  wait
  done=$(grep -l '^200$' /tmp/sw.* 2>/dev/null | wc -l | tr -d ' ')
  echo "$done $held"
}

read -r c1 h1 < <(run_round "$((POOL - 1))" "${TAG}below_" 20)
read -r c2 h2 < <(run_round "$POOL"         "${TAG}at_"    20)

printf '   pool=%-3s  %2d concurrent -> %2d/%-2d completed, %s idle-in-tx   |  %2d concurrent -> %2d/%-2d completed, %s idle-in-tx  %s\n' \
  "$POOL" "$((POOL - 1))" "$c1" "$((POOL - 1))" "$h1" "$POOL" "$c2" "$POOL" "$h2" \
  "$( [ "$c1" = "$((POOL - 1))" ] && [ "$c2" = "0" ] && [ "$h2" = "$POOL" ] && echo '<- threshold exact' || echo '<- MISMATCH' )"
