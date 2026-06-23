# Custom Project Roles — Key Changes

Summary of the changes on branch `origin/18124-custom-role-feature` (PR
[goharbor/harbor#22815](https://github.com/goharbor/harbor/pull/22815)),
measured against upstream merge-base `ee1177d33`: **94 files, +9,146 / −573**.

The feature lets system administrators define **custom project roles** with an
arbitrary set of permissions at runtime — no code changes or rebuild required —
while keeping the existing built-in roles immutable and the permission model
backward compatible.

---

## 1. Data model & migration

`make/migrations/postgresql/0190_2.16.0_schema.up.sql`

- Extends the existing `role` table with: `is_builtin`, `description`,
  `modified`, `created_by`, `created_at`, `modified_by`, `modified_at`.
- Marks the seeded roles (`projectAdmin`, `developer`, `maintainer`, `guest`,
  `limitedGuest`) as **built-in / immutable**.
- **Moves built-in role permissions from hardcoded Go (`rolePoliciesMap`) into
  the database** (`permission_policy` + `role_permission`). This is the
  cornerstone change: every role's permissions now live in one place, which is
  what makes both custom roles and the anti-escalation check possible.

Architecture:

```
users / groups → project_member → role
                                   ↓
                             role_permission → permission_policy
```

The `role_permission` table already existed for robot accounts; it is now also
used for project roles (discriminated by role type).

## 2. Backend — new role domain

New layered packages following Harbor's manager/controller/handler pattern:

| Layer | Path |
|-------|------|
| Persistence (DAO) | `src/pkg/role/dao/`, `src/pkg/role/manager.go`, `src/pkg/role/model/` |
| Business logic | `src/controller/role/controller.go`, `src/controller/role/model.go` |
| REST handler | `src/server/v2.0/handler/role.go` (~322 lines), `handler/model/role.go` |
| Test mocks | `src/testing/controller/role/`, `src/testing/pkg/role/` |

New REST API (`api/v2.0/swagger.yaml`):

- `GET  /roles` — list roles
- `POST /roles` — create custom role
- `GET  /roles/{role_id}` — get role
- `PUT  /roles/{role_id}` — update custom role
- `DELETE /roles/{role_id}` — delete custom role

## 3. Security

- **Anti-escalation on member role assignment** — a caller cannot grant a role
  whose permissions exceed their own (`handler/member.go`, `handler/permissions.go`,
  `controller/member`). This is why permissions had to move to the DB (§1): the
  check compares any target role against the caller's effective permissions.
- **Admin-only mutation** — only system admins can create/modify/delete roles.
- **Correct admin detection** — `SessionUser` check uses `has_admin_role`
  instead of the raw `sysadmin_flag`.
- **Built-in roles immutable** — enforced via `is_builtin`; the UI also disables
  edit/delete for them.
- **Audit logging** — role operations emit audit events
  (`src/controller/event/metadata/role.go`).
- RBAC plumbing updated in `src/common/rbac/`, `src/common/security/`,
  `src/common/models/role.go`.

## 4. Performance — permission cache

Implemented as a **decorator** (`cachingController` in
`src/controller/role/cache.go`) that implements the same `Controller` interface
and wraps the DB-backed `controller`. `role.Ctl` is the decorated instance; the
inner controller stays completely cache-agnostic. It is a **two-level cache**:

- **L1 — process-local** (`sync.Map` of `l1Entry`, per node): zero network hops
  on the hot path. Each entry carries an absolute expiry.
- **L2 — shared Redis** via `cache.Default()` (the same instance the quota
  controller uses), keyed by `role:<id>`.

Read path for `Get(WithPermission)`: `L1 (fresh) → L2 (Redis) → inner (DB)`;
results are written back to the enabled layers. `Get` without permissions, plus
`Count`/`List`, pass straight through to the inner controller (the latter two via
struct embedding). `Create`/`Update`/`Delete` delegate to the inner controller
and then invalidate the changed role in both layers; other nodes refresh once
their L1 entry expires (the shared L2 entry was deleted, so they re-read the DB).
This bounds cross-node staleness to the L1 window without any extra coordination —
the earlier Redis "version token" / generation machinery was removed as redundant.

**Configuration (env, read once at startup).** Values accept an integer number
of seconds or a Go duration string (e.g. `30m`); `<= 0` (i.e. `-1`) disables that
layer. This mirrors the parser in `src/lib/cache/redis/util.go`.

| Env var | Default | Effect |
|---|---|---|
| `ROLE_CACHE_L1_MEMORY_TTL` | `1s` | L1 freshness window; `-1` disables L1 (every read consults L2/DB). |
| `ROLE_CACHE_L2_REDIS_TTL` | `30m` | Redis entry TTL / backstop for out-of-band edits; `-1` disables Redis (reads go L1 → DB). |

Setting both to `-1` is the full-bypass mode (DB every request). There is no
separate on/off flag — the two TTLs express every mode.

Behaviour matrix:

| `..._L1_MEMORY_TTL` | `..._L2_REDIS_TTL` | Read path |
|---|---|---|
| `1s` | `30m` | L1 → L2 → DB |
| `-1` | `30m` | L2 → DB every request |
| `1s` | `-1` | L1 → DB (Redis off) |
| `-1` | `-1` | DB every request (full bypass) |

## 5. Frontend (Angular portal)

- **Roles management UI** under `left-side-nav/roles/`: list grid, `add-role`
  dialog, and a reusable `role-permissions-panel` component.
- **Member assignment UI**: dynamic role selector in `add-member` / `add-group`
  so custom roles can be assigned to users and groups.
- **i18n**: new translation keys added across all 9 shipped languages
  (en, de, es, fr, ko, pt-br, ru, tr, zh-cn, zh-tw).
- **Framework fixes** for the Angular 21 / Clarity 18 / Node 22 baseline:
  - `standalone: false` on the fork's role components (NG6008).
  - `roles.component.html`: restored datagrid selection checkboxes by adding
    `clrDgSelectionType="multi"` and migrating row iteration `*ngFor` → `@for`
    (the rest of the portal already used `@for`). Without this the admin could
    not select a role to use Edit/Delete.

## 6. Tests

- **Backend unit tests**: `role_test.go`, `member_test.go`, `robot_test.go`,
  controller/dao/manager tests, plus generated mocks.
- **Integration suite**: `tests/integration/test_custom_roles.py` (~653 lines)
  covering CRUD, assignment, and the anti-escalation guarantees.

---

## Reviewer notes

- The PR diff is large because the branch was rebased onto recent `upstream/main`
  (which pulled in the Angular 21 / Clarity 18 / Node 22 upgrade). The
  feature-specific surface is roles backend (§2), security (§3), cache (§4), and
  the portal roles UI (§5).
- Permission changes are **session-scoped**: a user's effective permissions
  refresh at next login, consistent with existing Harbor behavior.
- Backward compatible: built-in roles and existing project-member flows are
  unchanged for callers that don't use custom roles.
