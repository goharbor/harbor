# DB connection pool exhaustion: a reproduction

> **This directory reproduces a bug. It does not fix it.**
>
> Everything here exists to make [#23879](https://github.com/goharbor/harbor/issues/23879)
> observable on demand: a core that deadlocks on its own database pool and stays
> deadlocked until it is restarted. The defect it demonstrates is in production
> code and is still present on `main`. See [What still needs fixing](#what-still-needs-fixing).

A one-command, fully containerised reproduction of #23879: core deadlocks on its
own database pool when several `DELETE /api/v2.0/users/{id}` requests are in
flight at once, which is what the portal's bulk-delete button sends.

Nothing here is a mock. It builds core from the checkout and drives the real
HTTP API. The only setting moved away from a default is the pool size, and only
so the arithmetic is small enough to watch.

The deadlock needs a bounded pool, so what that bound is set to decides how
reachable it is. `POSTGRESQL_MAX_OPEN_CONNS` defaults to `0` in
`src/lib/config/metadata/metadatalist.go`, meaning unbounded, in which case the
pool cannot be what runs out; `make/harbor.yml.tmpl` ships `max_open_conns: 900`;
and the instance in #23879 was running 25. Any deployment that tunes the pool
down — or runs several replicas against one server and sizes each pool
accordingly — lands in range.

## Run it

```bash
make gen_apis                  # if src/server/v2.0/restapi is not generated yet
cd tests/dbpool-exhaustion

./run.sh                       # pool of 5
DBDEMO_POOL=25 ./run.sh        # the pool size from the incident
DBDEMO_KEEP=1 ./run.sh         # leave the stack up afterwards
./sweep.sh                     # measure the threshold across pool sizes
```

Requires Docker Compose v2 and about 10 GB of disk for the Go build cache. The
first run compiles core, which takes a minute or two; later runs reuse the
cache. If `asciinema` is installed the run is recorded to `dbpool-repro.cast`;
`DBDEMO_NO_REC=1` skips that.

Everything runs under the compose project `harbor-dbpool-repro` with no
published host ports, so it cannot collide with a Harbor install on the same
machine.

## What it shows

```
PHASE 0   ping/health/projects/systeminfo all 200
PHASE 1    4 concurrent deletes -> 4/4 HTTP 200, all inside 0.09s
PHASE 2   10 concurrent deletes -> 0 complete; 5 sessions idle in transaction
PHASE 3   unrelated GETs hang too; GET /api/v2.0/ping still 200 in 0.001s
PHASE 4   every client gives up at 45s; all 5 transactions still open
PHASE 5   nothing was deleted
```

Phase 3 is the part that turns a bug on one endpoint into an outage. A `GET`
skips the transaction middleware and touches none of the contended rows, but it
still needs one pooled connection, so it hangs with the rest. Meanwhile
`GET /api/v2.0/ping` answers in a millisecond, because `pingSkipper` keeps it
clear of every DB-touching middleware — so Kubernetes liveness and readiness
stay green on a core that is serving nothing.

Phase 4 is the part that makes it an incident rather than a stall. After every
client has hung up, all the transactions are still open. Nothing in the stack
can cancel them, so core only comes back by being restarted. `run.sh` restarts
it afterwards to show exactly that.

Core logs no error at all through any of this. Every wedged request gets as far
as the audit-log config read in `PreCheck` and simply stops.

## Why concurrent user deletes

`transaction.Middleware` opens a transaction, and so takes a pooled connection,
for every non-GET request, and holds it for the whole request
(`src/core/middlewares/middlewares.go`). Inside that span `log.Middleware` runs
`commonevent.Metadata.PreCheck` *before* calling the next handler
(`src/server/middleware/log/log.go`), and for a DELETE on a resource with
`ShouldResolveName` that calls `IDToNameFunc`
(`src/pkg/auditext/event/basic.go`):

```go
// src/pkg/auditext/event/user/user.go
// use different context to so that the user is visible before the transaction is committed
user, err := pkgUser.Mgr.Get(orm.Context(), int(id))
```

`orm.Context()` builds a fresh context with a *pooled* Ormer, so that lookup
takes a **second** connection from the same pool while the request's first one is
still held. `POSTGRESQL_MAX_OPEN_CONNS` such requests in flight at once is a
closed cycle: every connection is held by a request that needs one more.

Nothing bounds the wait. beego's `Ormer.Begin` is `BeginWithCtx(context.Background())`,
and its non-`WithCtx` query methods pass `context.Background()` too, so no
request deadline, client disconnect or shutdown signal reaches the pool.

Note that `notification.Middleware` sits *outside* `transaction.Middleware` by
design, so audit events publish after the commit. That protects `Resolve`. It
does not protect `PreCheck`, which is the one that runs inside the transaction.

The same shape exists at `src/pkg/oidc/helper.go` (`populateGroupsDB`, on every
OIDC-authenticated write), `src/pkg/authproxy/http.go`, `src/core/api/internal.go`
(`POST /api/internal/syncquota`) and `src/pkg/auditext/event/member/member.go`.

## The threshold

`./sweep.sh` restarts core at each pool size and fires real concurrent deletes,
with no barrier and no synthetic handler:

```
pool=5     4 concurrent ->  4/4  completed, 0 idle-in-tx  |   5 -> 0/5  completed,  5 idle-in-tx
pool=10    9 concurrent ->  9/9  completed, 0 idle-in-tx  |  10 -> 0/10 completed, 10 idle-in-tx
pool=25   24 concurrent -> 24/24 completed, 0 idle-in-tx  |  25 -> 0/25 completed, 25 idle-in-tx
```

The boundary is structural rather than statistical, and it does not need the
requests to be lined up artificially: a request holds its transaction connection
for its entire duration, so requests starting milliseconds apart are all still
holding connection #1 when the last one reaches for #2.

At pool 25 — the incident's configuration — twenty-four concurrent bulk deletes
are fine and the twenty-fifth takes the instance down.

## `pg_stat_activity` during the wedge

```
        state        | wait_event | sessions |  max_xact_age
---------------------+------------+----------+----------------
 idle in transaction | ClientRead |        5 | 00:00:08.00831
```

All sessions `idle in transaction` on `ClientRead`, no lock waits, PostgreSQL
idle and waiting on Harbor rather than the other way round. PostgreSQL is
started with `max_connections=200` against a core pool of 5, so the server can
never be what runs out.

## What still needs fixing

Nothing in this directory changes behaviour. The defect is in production code,
and a run of `./run.sh` on a patched tree should fail — phase 2 should complete
instead of wedging, or return an error instead of hanging forever.

Two changes are needed, and they are independent:

**1. A request must never hold two pooled connections.** Thread the request
context through the call sites that build their own: `IDToNameFunc` in
`src/pkg/auditext/event/user/user.go`, `ensureORMContext` in
`src/pkg/auditext/event/member/member.go`, `populateGroupsDB` in
`src/pkg/oidc/helper.go`, `src/pkg/authproxy/http.go`, and
`src/core/api/internal.go`. Where a value genuinely must be read outside the
transaction — `userIDToName`'s comment says the user must be visible before the
commit — the read wants deferring until after the commit rather than running on
a second connection during it; `commonevent.Metadata` is already carried to
`notification.Middleware`, which runs outside the transaction by design, so the
name could be resolved there instead of in `PreCheck`. This removes the specific
cycle in #23879.

**2. The wait for a connection must be bounded.** `ormerTx.Begin`
(`src/lib/orm/tx.go`) calls beego's `Ormer.Begin()`, which is
`BeginWithCtx(context.Background())`; routing it through `BeginWithCtx(ctx)`
instead would let a request deadline reach `BeginTx`. Beego's non-`WithCtx`
query methods pass `context.Background()` as well, so the same applies to reads
taken inside a request. This fixes no deadlock by itself, and that is
the point: it converts every one of them — the cycle above, the ones at the call
sites listed above, and the ones not found yet — from a total outage into slow
requests or errors on one endpoint.

The first without the second leaves the next such call site free to take the
whole instance down. The second without the first turns this outage into a burst
of failing requests, which is a large improvement on its own.

Worth considering alongside them: `GET /api/v2.0/ping` deliberately skips every
DB-touching middleware, which is correct for what that endpoint is for, but it
means a readiness probe cannot distinguish a healthy core from a wedged one.

## Files

| | |
|---|---|
| `docker-compose.yml` | PostgreSQL, Redis, registry, core built from the checkout, and a driver container |
| `driver.Dockerfile` | Alpine with `curl`, `psql` and `jq` |
| `scenario.sh` | The reproduction, run inside the driver container |
| `run.sh` | Brings the stack up, runs the scenario, shows the logs, restarts core |
| `sweep-scenario.sh` | One threshold measurement: `POOL-1` concurrent deletes, then `POOL` |
| `sweep.sh` | Repeats that across pool sizes |
