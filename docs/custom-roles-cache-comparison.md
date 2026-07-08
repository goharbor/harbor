# Custom Project Roles — Permission Caching: Options Considered

Custom roles move built-in role permissions out of compile-time Go constants
(`rbac_role.go`) and into the database (`role_permission` + `permission_policy`).
Project-member authorization is evaluated **on every request**, so each request now
risks an extra `role_permission` lookup. The question the reviewers debated was how
to avoid paying that query per request **without** weakening revocation timeliness
or adding fragile infrastructure.

Measured cost of the naive path (k6, 500 VUs, role-evaluating endpoints):
**+14–56 %** per-request latency vs upstream. Admins bypass the role evaluator, so
they are unaffected either way.

## The options

Using the taxonomy from the design discussion (`0` = no cache, `1.X` = session-scope
cache, `2.X` = system-scope cache). The criteria that decided it:

- **Request time** — added per-request latency on role-evaluating endpoints
  (measured vs upstream where available).
- **Stale window** — how long a role change can go unseen (cross-node propagation /
  revocation timeliness).
- **Redis impact** — does the approach put Redis on the authz hot path?

Two cache **scopes** frame the families:

- **1.X — session-scope:** permissions cached per user session. The variants differ
  only in *when a session refreshes* — re-login, a TTL, session invalidation, or a
  Redis version-key. Each cached copy serves a single session.
- **2.X — system-scope:** permissions cached once per node (or in Redis) and shared
  across all sessions and users — one role's entry serves everyone.

| # | Approach (proposer) | Request time | Stale window | Redis impact | Verdict |
|---|---|---|---|---|---|
| 0 | **DB every request** (baseline) | all auth **+14–56 %** | none (always fresh) | none | Correct but too slow as default |
| **1** | **session-scope cache** | | | | |
| 1.1 | **At login** (initial, Max) | session **~0** · basic **+14–56 %** | until session ends  | none | Withdrawn — long stale window |
| 1.2 | **L1 session + opt-in L2 Redis** (review) | session **~0** · basic **+14–56 %** | **= L1 TTL**; must be **large** (N sessions/node reload on expiry) | none by default; opt-in L2, but **N reads/node** on expiry (1/session) | Deferred — Harbor-wide perf follow-up |
| 1.3 | **Login + session invalidation** (Vad1mo) | session **~0** · basic **+14–56 %** | none (fresh or kickout) | (relogin spike) | Set aside — write-back race + reverse lookup |
| 1.4 | **Session cache + Redis version-key** (wy65701436) | session **~0** · basic **+14–56 %** | ~1 s (on version bump) | read-path: version-key per request (+ reload spike) | Rejected — Redis on authz path (#23335) |
| **2** | **system-scope cache** | | | | |
| 2.1 | **Redis-only shared cache** (considered) | all auth: **+1 Redis hop/request** | ~0 (delete on write) | read-path: GET every request | Not pursued — Redis on every request (cf. #19156/63/64) |
| 2.2 | **L1 node + opt-in L2 Redis** (chosen) | all auth **~0** (parity; 0 hops) | **= L1 TTL** (1s) | none by default; opt-in L2, **1 read/node** on expiry | **Implemented** |

> **Coverage of the session-scope family (1.X):** a per-session cache only speeds up
> session-backed auth (UI/OIDC-web). Basic auth carries no session to cache in
> (`local.NewSecurityContext` is rebuilt per request), so under any 1.X option it
> still pays the full per-request `role_permission` query (**+14–56 %**). The
> system-scope family (2.X) caches at the role→permission level, so it removes that
> cost for **every** auth path — a decisive advantage of 2.2.

> **Why session-scope can't just use a short TTL or instant invalidation:** 1.2 and
> 2.2 are the *same* design (L1 + optional L2 + TTL) — they differ only in **L1
> scope**. Because there are **many sessions per node**, a per-session L1 that
> reloads on TTL expiry multiplies the reload count by the session count, forcing
> 1.2 to a *large* stale window (where 2.2's per-node L1 reloads once).
> And invalidating every session of a modified role (1.3 / 1.4) makes a single change
> to a popular role trigger a **simultaneous reload across all those sessions**,
> spiking both DB and Redis. The system-scope cache (2.2) holds **one entry per role
> per node**, so its refresh (one read/role/node per TTL) and its invalidation (drop
> one entry/node) are both independent of how many users or sessions are active —
> which is why it can afford a `1 s` window cheaply.

## Why the others were not adopted

- **1.1 (initial proposal) rested on a wrong assumption.** It assumed project role
  was already frozen in the session, so caching permissions there would "match
  existing behavior." Testing (@chlins) showed project role is actually re-read from
  the DB **every request** — so 1.1 would have *regressed* today's immediate
  revocation. (A genuinely session-frozen field does exist — `GroupIDs` /
  `SysAdminFlag`, set once at login — but project role is not it.)

- **1.4 (session cache + Redis version-key)** caches the permissions per session and
  checks a Redis version stamp (`custom_roles:last_update`) during permission
  evaluation, lazily reloading when it changes. This gives instant cross-node
  invalidation but puts a Redis dependency on the most security-critical,
  highest-traffic path. As @Vad1mo noted, the version-key check "is still a
  round-trip, or a local cache of the version key, which just moves the staleness
  question up one level." He objected citing goharbor/harbor#23335 — a per-request
  Redis cache path (`FetchOrSave` → keyMutex) that exhausted the DB connection pool
  and hung core — and the repeated efforts to move config caching *off* Redis to
  memory (#19156 / #19163 / #19164). The chosen **2.2** caches the same permissions
  at system scope but invalidates by a **time-based L1 TTL** instead of a Redis
  signal, keeping Redis off the read path (and optional).

- **1.2 / 1.3** are attractive but larger in scope: 1.2 touches the security-context
  build path (and is a Harbor-wide win worth its own proposal); 1.3 needs a
  session→role reverse index that is awkward without shared session state, leans on
  Redis (the session store) on the write path, and only affects session-based auth
  (UI/OIDC-web). Non-session API auth such as basic auth builds a fresh security
  context per request (`local.NewSecurityContext` in
  `server/middleware/security/basic_auth.go`), so it has nothing to invalidate and no
  stale window to fix; robots are not role-based at all. Crucially 1.3 refreshes
  **only** on a write and has **no TTL backstop**. Cross-instance *propagation* is not
  the problem — the session blob lives in shared Redis and `SessionRead`
  (`core/session/session.go`) re-`Fetch`es it every request, so a rewritten blob is
  seen on the next request on any node, no broadcast needed. The problem is that you
  cannot *reliably* invalidate an ongoing session: every request writes the whole blob
  back on `SessionRelease` (`Set`, not `SetXX`), so an in-flight request that loaded
  the old values clobbers the external edit — and even a delete is resurrected. Add the
  reverse lookup (sessions are keyed by sid, not user), and anything invalidation
  misses stays stale until the session ends. 2.2's L1/L2 TTLs are the backstop 1.3
  lacks — a missed invalidation self-heals within the TTL.

## The chosen design (2.2)

A `cachingController` decorator (`src/controller/role/cache.go`) wraps a
cache-agnostic DB controller:

```
Get(WithPermission):  L1 (process memory, TTL) → [optional] L2 (Redis, TTL) → DB
Create/Update/Delete: write to DB, then delete L2 key + drop L1
```

It satisfies the agreed priority ladder — **(1) no Redis, (2) in-memory TTL, (3)
Redis only as an opt-in toggle that reverts to no-Redis when off**:

- **No Redis by default.** Out of the box only L1 memory is active
  (`ROLE_CACHE_L1_MEMORY_TTL`, default `1 s`). Redis (L2) is **off** unless
  `ROLE_CACHE_L2_REDIS_TTL` is set (e.g. `30m`). Either layer disables independently;
  both off = DB every request.
- **Off the hot path.** Per-request reads are served from L1 memory — zero network
  hops — so the authz path has no Redis availability/latency dependency.
- **Avoids the #23335 failure mode.** Uses plain `Fetch`/`Save`, **not**
  `FetchOrSave`, so it never takes the keyMutex path; a Redis error falls through to
  the DB rather than blocking.
- **No version-key.** Cross-node freshness is just the L1 TTL plus an L2-delete on
  write — bounded staleness with no shared-state round-trip.

**Trade-off:** a change made on one node is visible on others within the L1 TTL
(1 s default), not instantly. Direct out-of-band SQL edits are caught by the L2 TTL
when Redis is enabled. For a setting that changes once or twice in a role's
lifetime, this is an accepted bound.

Result: with the cache on, the role-evaluating endpoints return to **upstream
parity**; with it off (or unavailable) the feature degrades to the correct DB path.
