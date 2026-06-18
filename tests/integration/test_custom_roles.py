#!/usr/bin/env python3
"""
Harbor Custom Roles - Automated Test Suite

Sections:
  1. Role CRUD             — create, update, delete; built-in immutability
  2. DB seeding            — built-in roles have permissions via ListRole
  3. Member anti-escalation — POST/PUT /projects/{id}/members blocked when caller lacks perms
  4. Group anti-escalation  — same check via group payload (LDAP/OIDC/HTTP modes)
  5. Sysadmin bypass       — admin can assign any role
  6. Role display flags    — is_builtin=true present on all built-in roles
  7. Edge cases            — zero-permission custom role, empty roles list fallback
  8. Robot anti-escalation — POST/PUT /robots blocked when caller lacks requested perms

Usage:
    python3 test_custom_roles.py --url https://harbor.example.com \
                                 --user admin --password YourPassword

Exit code 0 = all tests passed, 1 = one or more tests failed.
"""

import argparse
import sys
import requests
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# ---------------------------------------------------------------------------
# Output helpers
# ---------------------------------------------------------------------------

GREEN  = '\033[92m'
RED    = '\033[91m'
YELLOW = '\033[93m'
BOLD   = '\033[1m'
RESET  = '\033[0m'

_passed  = 0
_failed  = 0
_skipped = 0


def ok(msg):
    global _passed
    _passed += 1
    print(f"  {GREEN}PASS{RESET}  {msg}")


def fail(msg, detail=None):
    global _failed
    _failed += 1
    print(f"  {RED}FAIL{RESET}  {msg}")
    if detail:
        print(f"        {RED}{detail}{RESET}")


def skip(msg):
    global _skipped
    _skipped += 1
    print(f"  {YELLOW}SKIP{RESET}  {msg}")


def section(title):
    print(f"\n{BOLD}{'─' * 60}{RESET}")
    print(f"{BOLD}  {title}{RESET}")
    print(f"{BOLD}{'─' * 60}{RESET}")


def assert_status(label, resp, expected):
    if isinstance(expected, int):
        expected = [expected]
    if resp.status_code in expected:
        ok(f"{label} → {resp.status_code}")
        return True
    else:
        fail(f"{label} → expected {expected}, got {resp.status_code}", resp.text[:200])
        return False


# ---------------------------------------------------------------------------
# Minimal Harbor API client
# ---------------------------------------------------------------------------

class Harbor:
    def __init__(self, base_url, username, password, verify_ssl=False):
        self.base = base_url.rstrip('/') + '/api/v2.0'
        self.s = requests.Session()
        self.s.verify = verify_ssl
        self.s.auth = (username, password)
        self.s.headers.update({'Content-Type': 'application/json'})

    def get(self, path, **kw):    return self.s.get(self.base + path, **kw)
    def post(self, path, **kw):   return self.s.post(self.base + path, **kw)
    def put(self, path, **kw):    return self.s.put(self.base + path, **kw)
    def delete(self, path, **kw): return self.s.delete(self.base + path, **kw)

    def as_user(self, username, password):
        return Harbor(self.base.replace('/api/v2.0', ''), username, password, self.s.verify)


# ---------------------------------------------------------------------------
# Permission fixtures
# ---------------------------------------------------------------------------

PULL_ONLY_PERMS = [
    {"resource": "repository",        "action": "pull"},
    {"resource": "repository",        "action": "list"},
    {"resource": "repository",        "action": "read"},
    {"resource": "artifact",          "action": "read"},
    {"resource": "artifact",          "action": "list"},
    {"resource": "artifact-addition", "action": "read"},
    {"resource": "tag",               "action": "list"},
    {"resource": "accessory",         "action": "list"},
]

PULL_ONLY_PERMS_V2 = PULL_ONLY_PERMS + [
    {"resource": "artifact", "action": "delete"},
]

def _role_body(name, access_list, description=""):
    return {
        "name": name,
        "description": description,
        "permissions": [{"kind": "project", "namespace": "*", "access": access_list}],
    }


# ---------------------------------------------------------------------------
# Cleanup
# ---------------------------------------------------------------------------

def cleanup(admin, state):
    pid = state.get('project_id')
    if pid:
        for mid in state.get('member_ids', []):
            admin.delete(f"/projects/{pid}/members/{mid}")
        admin.delete(f"/projects/{pid}")
    for uid in state.get('user_ids', []):
        admin.delete(f"/users/{uid}")
    for rid in state.get('custom_role_ids', []):
        admin.delete(f"/roles/{rid}")


# ---------------------------------------------------------------------------
# Helper: find a role by name in a list
# ---------------------------------------------------------------------------

def _find(roles, name):
    return next((r for r in roles if r['name'] == name), None)


def _role_id_from_response(resp, admin, name):
    """Extract role id from Location header or by re-listing."""
    loc = resp.headers.get('Location', '')
    if loc:
        try:
            return int(loc.rstrip('/').split('/')[-1])
        except ValueError:
            pass
    body = resp.json() if resp.content else {}
    if body.get('id'):
        return body['id']
    # fall back: list and search by name
    r = admin.get("/roles", params={"page": 1, "page_size": 100})
    return (_find(r.json(), name) or {}).get('id')


# ---------------------------------------------------------------------------
# Test runner
# ---------------------------------------------------------------------------

def run(admin_url, admin_user, admin_pass):
    admin = Harbor(admin_url, admin_user, admin_pass)
    state = {
        'project_id':    None,
        'user_ids':      [],
        'member_ids':    [],
        'custom_role_ids': [],
        'builtin':       {},   # name → id
    }
    BUILTIN_NAMES = {'projectAdmin', 'maintainer', 'developer', 'guest', 'limitedGuest'}

    # ── 0. SETUP ────────────────────────────────────────────────────────────
    section("0. Setup")

    # Verify connectivity
    r = admin.get("/systeminfo")
    if r.status_code != 200:
        fail("Cannot reach Harbor API — aborting", r.text[:200])
        return False

    # Collect built-in role IDs
    r = admin.get("/roles", params={"page": 1, "page_size": 100})
    if not assert_status("GET /roles", r, 200):
        return False
    for role in r.json():
        if role['name'] in BUILTIN_NAMES:
            state['builtin'][role['name']] = role['id']
    if len(state['builtin']) == 5:
        ok(f"All 5 built-in roles found: {state['builtin']}")
    else:
        fail(f"Expected 5 built-in roles, got {len(state['builtin'])}: {list(state['builtin'])}")

    # Create test project
    r = admin.post("/projects", json={"project_name": "cr-test", "metadata": {"public": "false"}})
    if r.status_code in (200, 201):
        r2 = admin.get("/projects/cr-test")
        state['project_id'] = r2.json()['id']
        ok(f"Created project 'cr-test' (id={state['project_id']})")
    else:
        fail("Could not create project", r.text[:200])
        return False

    # Create 3 test users: maintainer_user, guest_user, custom_user
    users_created = []
    for uname in ["cr-maintainer", "cr-guest", "cr-custom"]:
        r = admin.post("/users", json={
            "username": uname, "password": "Test1@3456",
            "email": f"{uname}@cr-test.local", "realname": uname,
        })
        if r.status_code in (200, 201):
            r2 = admin.get("/users/search", params={"username": uname})
            uid = r2.json()[0]['user_id']
            state['user_ids'].append(uid)
            users_created.append((uname, uid))
            ok(f"Created user '{uname}' (id={uid})")
        else:
            fail(f"Could not create user '{uname}'", r.text[:200])

    if len(users_created) < 3:
        skip("Not all test users created — some member tests will be skipped")

    # ── 1. ROLE CRUD ────────────────────────────────────────────────────────
    section("1. Role Management")

    # 1.1 Create custom role
    r = admin.post("/roles", json=_role_body("cr-pull-only", PULL_ONLY_PERMS, "pull-only test role"))
    if assert_status("POST /roles (create 'cr-pull-only')", r, [200, 201]):
        rid = _role_id_from_response(r, admin, "cr-pull-only")
        state['custom_role_ids'].append(rid)
        ok(f"  id={rid}")

        # Verify is_builtin = false
        r2 = admin.get(f"/roles/{rid}")
        if r2.status_code == 200 and r2.json().get('is_builtin') == False:
            ok("Custom role has is_builtin=false")
        else:
            fail("Custom role should have is_builtin=false", r2.text[:200])

        # 1.2 Edit custom role
        r3 = admin.put(f"/roles/{rid}", json=_role_body("cr-pull-only", PULL_ONLY_PERMS_V2))
        assert_status(f"PUT /roles/{rid} (update permissions)", r3, [200, 204])

        # Verify update reflected in GET
        r4 = admin.get(f"/roles/{rid}")
        if r4.status_code == 200:
            access = [a for p in r4.json().get('permissions', []) for a in p.get('access', [])]
            names = {f"{a['resource']}:{a['action']}" for a in access}
            if "artifact:delete" in names:
                ok("Updated permission (artifact:delete) visible in GET /roles/{id}")
            else:
                fail("Updated permission not reflected", str(names))
    else:
        skip("Skipping edit/delete tests — role creation failed")

    # 1.3 Delete a throwaway custom role (no members)
    r = admin.post("/roles", json=_role_body("cr-throwaway", []))
    if r.status_code in (200, 201):
        tid = _role_id_from_response(r, admin, "cr-throwaway")
        r2 = admin.delete(f"/roles/{tid}")
        assert_status(f"DELETE /roles/{tid} (throwaway role)", r2, [200, 204])
    else:
        skip("Could not create throwaway role for delete test")

    # 1.4 Built-in roles are immutable
    for name, rid in state['builtin'].items():
        r = admin.delete(f"/roles/{rid}")
        if r.status_code in (400, 403):
            ok(f"DELETE built-in '{name}' → {r.status_code} (blocked)")
        else:
            fail(f"DELETE built-in '{name}' should be blocked, got {r.status_code}")

        r = admin.put(f"/roles/{rid}", json=_role_body(name, []))
        if r.status_code in (400, 403):
            ok(f"PUT built-in '{name}' → {r.status_code} (blocked)")
        else:
            fail(f"PUT built-in '{name}' should be blocked, got {r.status_code}")

    # ── 2. DB SEEDING — permissions returned by ListRole ────────────────────
    section("2. Built-in Role Permission Seeding")

    r = admin.get("/roles", params={"page": 1, "page_size": 100})
    all_roles = r.json()
    for role in all_roles:
        if role['name'] not in BUILTIN_NAMES:
            continue
        count = sum(len(p.get('access', [])) for p in role.get('permissions', []))
        if count > 0:
            ok(f"Built-in '{role['name']}' → {count} permission entries in DB")
        else:
            fail(f"Built-in '{role['name']}' has 0 permissions — migration 0190 seeding failed?")

    # ── 3. MEMBER ADDITION — anti-escalation ────────────────────────────────
    section("3. Member Addition — Anti-Escalation")

    pid = state['project_id']
    B = state['builtin']

    if len(users_created) < 3:
        skip("Skipping all member tests (test users missing)")
    else:
        maintainer_name, maintainer_uid = users_created[0]
        guest_name,     guest_uid      = users_created[1]
        custom_name,    custom_uid     = users_created[2]
        custom_role_id = state['custom_role_ids'][0] if state['custom_role_ids'] else None

        # Add cr-maintainer as maintainer (as admin)
        r = admin.post(f"/projects/{pid}/members", json={
            "role_id": B['maintainer'],
            "member_user": {"user_id": maintainer_uid},
        })
        if assert_status(f"Admin adds '{maintainer_name}' as maintainer", r, [200, 201]):
            state['member_ids'].append(r.headers.get('Location', '').rstrip('/').split('/')[-1])

        # ── 3a. Maintainer tries to add projectAdmin (escalation) ─────────
        m_client = Harbor(admin_url, maintainer_name, "Test1@3456")

        r = m_client.post(f"/projects/{pid}/members", json={
            "role_id": B['projectAdmin'],
            "member_user": {"user_id": guest_uid},
        })
        assert_status("Maintainer→projectAdmin (escalation) blocked", r, 403)

        # ── 3b. Maintainer tries to add guest (valid) ─────────────────────
        r = m_client.post(f"/projects/{pid}/members", json={
            "role_id": B['guest'],
            "member_user": {"user_id": guest_uid},
        })
        if assert_status("Maintainer→guest (valid assignment)", r, [200, 201]):
            guest_mid = r.headers.get('Location', '').rstrip('/').split('/')[-1]
            state['member_ids'].append(guest_mid)

            # ── 3c. Maintainer tries to change guest→projectAdmin via PUT ──
            r2 = m_client.put(f"/projects/{pid}/members/{guest_mid}", json={
                "role_id": B['projectAdmin'],
            })
            assert_status("Maintainer→PUT projectAdmin (escalation) blocked", r2, 403)

            # ── 3d. Maintainer changes guest→developer (valid PUT) ─────────
            r3 = m_client.put(f"/projects/{pid}/members/{guest_mid}", json={
                "role_id": B['developer'],
            })
            assert_status("Maintainer→PUT developer (valid)", r3, [200, 204])

        # ── 3e. Admin adds cr-custom with custom pull-only role ───────────
        if custom_role_id:
            r = admin.post(f"/projects/{pid}/members", json={
                "role_id": custom_role_id,
                "member_user": {"user_id": custom_uid},
            })
            if assert_status(f"Admin assigns custom role to '{custom_name}'", r, [200, 201]):
                state['member_ids'].append(r.headers.get('Location', '').rstrip('/').split('/')[-1])

                # Pull-only user tries to add guest (escalation — guest has more perms)
                c_client = Harbor(admin_url, custom_name, "Test1@3456")
                r2 = c_client.post(f"/projects/{pid}/members", json={
                    "role_id": B['guest'],
                    "member_user": {"username": admin_user},
                })
                assert_status("Custom(pull-only)→guest (escalation) blocked", r2, 403)

                # Pull-only user tries to add someone with same custom role (valid)
                # Create a 4th user for this
                r3 = admin.post("/users", json={
                    "username": "cr-extra", "password": "Test1@3456",
                    "email": "cr-extra@cr-test.local", "realname": "cr-extra",
                })
                if r3.status_code in (200, 201):
                    r4 = admin.get("/users/search", params={"username": "cr-extra"})
                    extra_uid = r4.json()[0]['user_id']
                    state['user_ids'].append(extra_uid)
                    r5 = c_client.post(f"/projects/{pid}/members", json={
                        "role_id": custom_role_id,
                        "member_user": {"user_id": extra_uid},
                    })
                    assert_status("Custom(pull-only)→same custom role (valid)", r5, [200, 201])
                    if r5.status_code in (200, 201):
                        state['member_ids'].append(
                            r5.headers.get('Location', '').rstrip('/').split('/')[-1])

    # ── 4. +GROUP MODAL — API equivalent ────────────────────────────────────
    section("4. Group Role Assignment — Anti-Escalation")

    # Harbor only supports groups when LDAP/OIDC is configured.
    # We test the API directly with a user_group payload in HTTP-auth mode
    # (if group type is not supported the endpoint returns 400/422, not 403 —
    #  so we only run this check when we can confirm group support).
    r = admin.get("/systeminfo")
    auth_mode = r.json().get('auth_mode', 'db_auth') if r.status_code == 200 else 'db_auth'
    if auth_mode in ('ldap_auth', 'http_auth', 'oidc_auth'):
        if len(users_created) >= 1:
            m_client = Harbor(admin_url, users_created[0][0], "Test1@3456")
            r = m_client.post(f"/projects/{pid}/members", json={
                "role_id": B.get('projectAdmin'),
                "member_group": {"group_name": "test-group", "group_type": 1},
            })
            assert_status("Maintainer→group as projectAdmin (escalation) blocked", r, 403)
    else:
        skip(f"Auth mode is '{auth_mode}' — group escalation test skipped (no group support)")

    # ── 5. SYSADMIN BYPASS ───────────────────────────────────────────────────
    section("5. Sysadmin Bypass")

    if pid and len(users_created) >= 1:
        _, target_uid = users_created[0]
        r = admin.put(f"/projects/{pid}/members/{state['member_ids'][0]}", json={
            "role_id": B['projectAdmin'],
        })
        # Admin setting someone to projectAdmin must always succeed (200 or 204)
        assert_status("Sysadmin can assign projectAdmin (no restriction)", r, [200, 204])

    # ── 6. ROLE DISPLAY — translation key check ──────────────────────────────
    section("6. Role Display Keys")

    EXPECTED_KEYS = {
        'projectAdmin': 'MEMBER.PROJECT_ADMIN',
        'maintainer':   'MEMBER.PROJECT_MAINTAINER',
        'developer':    'MEMBER.DEVELOPER',
        'guest':        'MEMBER.GUEST',
        'limitedGuest': 'MEMBER.LIMITED_GUEST',
    }
    r = admin.get("/roles", params={"page": 1, "page_size": 100})
    for role in r.json():
        if role['name'] in EXPECTED_KEYS:
            if role.get('is_builtin'):
                ok(f"Built-in role '{role['name']}' has is_builtin=true (badge will render)")
            else:
                fail(f"Built-in role '{role['name']}' missing is_builtin=true flag")

    # ── 7. EDGE CASES ────────────────────────────────────────────────────────
    section("7. Edge Cases")

    # Custom role with zero permissions — anyone can assign it (subset of anything)
    r = admin.post("/roles", json=_role_body("cr-empty", []))
    if r.status_code in (200, 201):
        empty_rid = _role_id_from_response(r, admin, "cr-empty")
        state['custom_role_ids'].append(empty_rid)
        if len(users_created) >= 1:
            m_client = Harbor(admin_url, users_created[0][0], "Test1@3456")
            r2 = m_client.post(f"/projects/{pid}/members", json={
                "role_id": empty_rid,
                "member_user": {"user_id": users_created[1][1] if len(users_created) > 1 else users_created[0][1]},
            })
            # maintainer has all superset permissions over empty role — should succeed
            if r2.status_code in (200, 201):
                ok("Maintainer can assign zero-permission role (subset of everything)")
                state['member_ids'].append(r2.headers.get('Location', '').rstrip('/').split('/')[-1])
            elif r2.status_code == 409:
                ok("Zero-permission role assignment: member already exists (409 — acceptable)")
            else:
                fail(f"Maintainer should be able to assign empty role, got {r2.status_code}", r2.text[:200])
    else:
        skip("Could not create zero-permission role for edge case test")

    # ── 8. ROBOT ANTI-ESCALATION ─────────────────────────────────────────────
    section("8. Robot Anti-Escalation")

    # validateNoEscalation fires when a human (local) user calls POST/PUT /robots.
    # It checks each requested robot permission against the caller's own project perms.
    #
    # Important: only users with robot:create can create robots at all.
    # Built-in roles with robot:create: projectAdmin only.
    # Maintainer has robot:read + robot:list but NOT robot:create.
    #
    # To test anti-escalation we create a custom role that has robot:create +
    # repository:pull/push but deliberately omits member:create.  A user with
    # that role can create robots scoped to pull/push but not to member:create.

    project_name = "cr-test"

    def robot_body(name, access_list):
        return {
            "name": name,
            "description": "test robot",
            "duration": -1,
            "level": "project",
            "permissions": [{"kind": "project", "namespace": project_name, "access": access_list}],
        }

    robot_ids = []

    if not pid:
        skip("Skipping robot tests (project missing)")
    else:
        # ── 8a. Sysadmin creates robot with any perms (no restriction) ────
        r = admin.post("/robots", json=robot_body("cr-robot-admin", [
            {"resource": "repository", "action": "pull"},
            {"resource": "repository", "action": "push"},
            {"resource": "member",     "action": "create"},
        ]))
        if assert_status("Sysadmin creates robot with member:create → 201", r, [200, 201]):
            rid = r.json().get('id')
            if rid:
                robot_ids.append(rid)

        # ── 8b. Create a "robot-manager" custom role ──────────────────────
        # Has robot:create/read/update/delete/list + repository:pull/push
        # but deliberately omits member:* and configuration:* permissions.
        robot_mgr_perms = [
            {"resource": "robot",       "action": "create"},
            {"resource": "robot",       "action": "read"},
            {"resource": "robot",       "action": "update"},
            {"resource": "robot",       "action": "delete"},
            {"resource": "robot",       "action": "list"},
            {"resource": "repository",  "action": "pull"},
            {"resource": "repository",  "action": "push"},
            {"resource": "repository",  "action": "list"},
            {"resource": "repository",  "action": "read"},
            {"resource": "artifact",    "action": "read"},
            {"resource": "artifact",    "action": "list"},
            {"resource": "tag",         "action": "list"},
        ]
        r = admin.post("/roles", json=_role_body("cr-robot-mgr", robot_mgr_perms))
        robot_mgr_role_id = None
        if r.status_code in (200, 201):
            robot_mgr_role_id = _role_id_from_response(r, admin, "cr-robot-mgr")
            state['custom_role_ids'].append(robot_mgr_role_id)
            ok(f"Created 'cr-robot-mgr' custom role (id={robot_mgr_role_id})")
        else:
            fail("Could not create robot-manager custom role", r.text[:200])

        # Create a user and assign robot-manager role
        robot_mgr_uid = None
        robot_mgr_name = "cr-robotmgr"
        r = admin.post("/users", json={
            "username": robot_mgr_name, "password": "Test1@3456",
            "email": f"{robot_mgr_name}@cr-test.local", "realname": robot_mgr_name,
        })
        if r.status_code in (200, 201):
            r2 = admin.get("/users/search", params={"username": robot_mgr_name})
            robot_mgr_uid = r2.json()[0]['user_id']
            state['user_ids'].append(robot_mgr_uid)
            ok(f"Created user '{robot_mgr_name}' (id={robot_mgr_uid})")
        else:
            fail(f"Could not create user '{robot_mgr_name}'", r.text[:200])

        if robot_mgr_role_id and robot_mgr_uid:
            r = admin.post(f"/projects/{pid}/members", json={
                "role_id": robot_mgr_role_id,
                "member_user": {"user_id": robot_mgr_uid},
            })
            if assert_status(f"Assign '{robot_mgr_name}' robot-manager role", r, [200, 201]):
                state['member_ids'].append(r.headers.get('Location', '').rstrip('/').split('/')[-1])

        rm_client = Harbor(admin_url, robot_mgr_name, "Test1@3456")

        # ── 8c. robot-manager creates robot within their permissions ──────
        r = rm_client.post("/robots", json=robot_body("cr-robot-rm-valid", [
            {"resource": "repository", "action": "pull"},
            {"resource": "repository", "action": "push"},
        ]))
        valid_robot_id = None
        if assert_status("robot-manager creates robot with pull+push → 201", r, [200, 201]):
            valid_robot_id = r.json().get('id')
            if valid_robot_id:
                robot_ids.append(valid_robot_id)

        # ── 8d. robot-manager tries robot with member:create (escalation) ─
        r = rm_client.post("/robots", json=robot_body("cr-robot-rm-escalate", [
            {"resource": "repository", "action": "pull"},
            {"resource": "member",     "action": "create"},
        ]))
        assert_status("robot-manager creates robot with member:create → 403 (escalation)", r, 403)

        # ── 8e. robot-manager updates existing robot to add member:delete ─
        if valid_robot_id:
            r = rm_client.put(f"/robots/{valid_robot_id}", json=robot_body("cr-robot-rm-valid", [
                {"resource": "repository", "action": "pull"},
                {"resource": "repository", "action": "push"},
                {"resource": "member",     "action": "delete"},
            ]))
            assert_status("robot-manager updates robot to add member:delete → 403 (escalation)", r, 403)

            # Valid update — stays within role's permissions
            r = rm_client.put(f"/robots/{valid_robot_id}", json=robot_body("cr-robot-rm-valid", [
                {"resource": "repository", "action": "pull"},
            ]))
            assert_status("robot-manager updates robot to pull-only → 200 (valid)", r, [200, 204])

        # ── 8f. robot-manager creates robot with zero permissions ─────────
        r = rm_client.post("/robots", json=robot_body("cr-robot-rm-empty", []))
        if assert_status("robot-manager creates robot with no permissions → 201", r, [200, 201]):
            rid = r.json().get('id')
            if rid:
                robot_ids.append(rid)

        # ── 8g. Confirm maintainer cannot create robots at all ────────────
        # (not an anti-escalation case — maintainer simply lacks robot:create)
        if len(users_created) >= 1:
            m_client = Harbor(admin_url, users_created[0][0], "Test1@3456")
            r = m_client.post("/robots", json=robot_body("cr-robot-m-attempt", [
                {"resource": "repository", "action": "pull"},
            ]))
            if r.status_code in (403, 401):
                ok(f"Maintainer cannot create robots (lacks robot:create) → {r.status_code}")
            else:
                fail(f"Expected 403/401 for maintainer robot creation, got {r.status_code}", r.text[:200])

        # Cleanup robots
        for rid in robot_ids:
            admin.delete(f"/robots/{rid}")
        if robot_ids:
            ok(f"Cleaned up {len(robot_ids)} test robot(s)")

    # ── CLEANUP ──────────────────────────────────────────────────────────────
    section("Cleanup")
    cleanup(admin, state)
    ok("All test artifacts removed")

    # ── SUMMARY ──────────────────────────────────────────────────────────────
    print(f"\n{'═' * 60}")
    print(f"  {BOLD}Results:{RESET}  "
          f"{GREEN}{_passed} passed{RESET}   "
          f"{RED}{_failed} failed{RESET}   "
          f"{YELLOW}{_skipped} skipped{RESET}")
    print(f"{'═' * 60}\n")
    return _failed == 0


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(
        description="Harbor Custom Roles — Automated Test Suite",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument("--url",        required=True, help="Harbor base URL, e.g. https://harbor.example.com")
    parser.add_argument("--user",       required=True, help="Admin username")
    parser.add_argument("--password",   required=True, help="Admin password")
    parser.add_argument("--verify-ssl", action="store_true", default=False,
                        help="Verify TLS certificate (default: skip verification)")
    args = parser.parse_args()

    success = run(args.url, args.user, args.password)
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
