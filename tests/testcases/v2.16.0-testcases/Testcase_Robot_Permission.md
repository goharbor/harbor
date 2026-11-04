**Description:** Verify backward compatibility of robot permissions for adding project members and changing their roles in Harbor v2.16.0. Cover built-in and custom roles, project scope, missing permissions, upgrade compatibility, and privilege escalation prevention for human users with custom roles.

**Reference:** User-provided screenshot of Harbor v2.15.3 versus v2.16.0-rc1 robot permission findings; [API specification](../../../api/v2.0/swagger.yaml); [Project Custom Role testcase](Testcase_Project_Custom_Role.md).

**Expected results from the reported scenarios**

The v2.15.3 and v2.16.0-rc1 columns below record the supplied findings, not results of this testcase. Successful robot member management must remain compatible with v2.15.3. The two human privilege escalation attempts must remain forbidden.

| Check | Reported v2.15.3 | Reported v2.16.0-rc1 | Expected v2.16.0 |
| --- | --- | --- | --- |
| Member-only and full robots add a member with any built-in role | 201 | 403 | 201 |
| Full robot changes Guest to Developer | 200 | 403 | 200 |
| Robot changes Guest to Project Admin | 200 | 403 | 200 |
| Human with only `member:update` promotes self to Project Admin | Custom roles unavailable | 403 | 403 |
| Human with only `robot:create/read/list` creates a robot with permissions the human does not hold | Custom roles unavailable | 403 | 403 |

**Steps**

Use Basic authentication as the actor specified in each step, without browser session cookies. Send JSON bodies with `Content-Type: application/json`. Replace host, project, robot, role, and member placeholders with actual values. Member IDs identify project memberships, not Harbor user IDs. Perform setup, independent read-back checks, fixture resets, and cleanup as the system administrator unless specified otherwise.

1. **Environment and robot fixtures**

   1. Prepare Harbor v2.15.3 as the compatibility baseline and a v2.16.0 candidate with custom roles enabled. Use database authentication and record the exact builds. Run the built-in-role checks in categories 2 and 3 on both versions. Run the remaining permission checks on v2.16.0. Include both a fresh v2.16.0 installation and an upgrade from the baseline as described in category 7.
   2. Create private projects `robot-permission-a` (A) and `robot-permission-b` (B). Create ordinary users `robot-add-target`, `robot-update-target`, `member-update-user`, and `robot-create-user`. Give none system administrator privileges or additional direct/group memberships. Add `robot-update-target` to both projects as Guest. Leave `robot-add-target` without project membership. Keep the system administrator as project owner throughout the test.
   3. As the system administrator, call `GET /api/v2.0/permissions` and save the response as `permissions.json`. Create the following robots with `POST /api/v2.0/robots`. All grants must have `effect: "allow"`, `kind: "project"`, and namespace `robot-permission-a` only. Do not grant system permissions or use the all-projects namespace `"*"`.

      | Robot fixture | Explicit access in project A |
      | --- | --- |
      | `member-only` | `member:create`, `member:read`, `member:list`, `member:update`, `member:delete` |
      | `full` | Every resource/action pair in the `project` array of `GET /api/v2.0/permissions` |
      | `member-reader` | `member:read`, `member:list` only |

      Save this body as `member-only.json`:

      ```json
      {
        "name": "member-only",
        "description": "Robot member permission regression test",
        "level": "project",
        "duration": 30,
        "disable": false,
        "permissions": [
          {
            "kind": "project",
            "namespace": "robot-permission-a",
            "access": [
              {"resource": "member", "action": "create", "effect": "allow"},
              {"resource": "member", "action": "read", "effect": "allow"},
              {"resource": "member", "action": "list", "effect": "allow"},
              {"resource": "member", "action": "update", "effect": "allow"},
              {"resource": "member", "action": "delete", "effect": "allow"}
            ]
          }
        ]
      }
      ```

      Generate the other bodies from the catalog and this template:

      ```bash
      jq --slurpfile catalog permissions.json \
        '.name = "full" | .permissions[0].access =
          ($catalog[0].project | map({resource, action, effect: "allow"}))' \
        member-only.json > full.json
      jq '.name = "member-reader" |
        .permissions[0].access |= map(select(.action == "read" or .action == "list"))' \
        member-only.json > member-reader.json
      ```

   4. Expect HTTP 201 for each robot creation. Record the returned ID, complete robot name, secret, and explicit grants. Use the returned name as the Basic-auth username; preserve its prefix and literal `$` character. Verify each robot can list members of A with HTTP 200 before testing writes. Keep secrets out of test reports.
   5. Repeat the fixtures and robot checks for system-level robots restricted to A: change `level` to `"system"` and use unique names such as `member-only-system`, keeping the same project permission block. A system-level robot in this testcase still has no grants in B. On v2.16.0, confirm built-in role IDs with `GET /api/v2.0/roles`: Project Admin `1`, Developer `2`, Guest `3`, Maintainer `4`, and Limited Guest `5`.

2. **Robots add members with built-in roles**

   1. For each combination of `member-only` and `full` robot, project/system robot level, and each built-in role below, ensure `robot-add-target` is not a member of A. Send `POST /api/v2.0/projects/robot-permission-a/members` as that robot, using the indicated numeric `role_id`.

      | Built-in role | `role_id` | Expected POST status |
      | --- | --- | --- |
      | Project Admin | 1 | 201 |
      | Developer | 2 | 201 |
      | Guest | 3 | 201 |
      | Maintainer | 4 | 201 |
      | Limited Guest | 5 | 201 |

      Save the body as `member-add.json`, changing `role_id` for each row:

      ```json
      {"role_id": 1, "member_user": {"username": "robot-add-target"}}
      ```

      Example request; `--user` prompts for the robot secret:

      ```bash
      curl -i --user '<complete-robot-name>' --request POST \
        'https://<harbor-host>/api/v2.0/projects/robot-permission-a/members' \
        --header 'Content-Type: application/json' \
        --data-binary @member-add.json
      ```

   2. After each success, follow the response's `Location` with an administrator GET. Verify the correct user, project, `role_id`, and `role_name`, and confirm the role in the project's Members page. Record the membership ID. Delete that membership as administrator before the next combination so a duplicate-member conflict cannot mask the authorization result.
   3. Record any HTTP 403 for these valid requests as a compatibility regression, including the response error code/message and robot grants. A member-only robot must be able to delegate built-in roles through its granted member management permission even though it cannot itself perform every action in those roles.

3. **Robots change an existing member's built-in role**

   1. Obtain `robot-update-target`'s membership ID in A. Before each attempt, reset this member to Guest (`role_id: 3`) as administrator and verify the reset. As each `member-only` and `full` robot at both robot levels, send `PUT /api/v2.0/projects/robot-permission-a/members/<member-id>` with each body below in a separate attempt.

      | Transition | JSON body | Expected PUT status |
      | --- | --- | --- |
      | Guest → Developer | `{"role_id":2}` | 200 |
      | Guest → Project Admin | `{"role_id":1}` | 200 |
      | Guest → Maintainer | `{"role_id":4}` | 200 |
      | Guest → Limited Guest | `{"role_id":5}` | 200 |

   2. After each PUT, retrieve the membership as administrator and refresh the Members page. Verify the requested role persisted. As the same robot, change the member back to Guest using `{"role_id":3}`; expect HTTP 200 and confirm the saved role again. Treat a 403 on an allowed transition as a regression even if custom-role updates work.

4. **Custom-role compatibility and human self-promotion protection**

   1. On v2.16.0, create custom role `Robot Test Member Updater` as administrator with `POST /api/v2.0/roles` and the following body. Expect HTTP 201 and record the role ID as `<member-update-role-id>`.

      ```json
      {
        "name": "Robot Test Member Updater",
        "description": "Only project member update permission",
        "permissions": [
          {
            "kind": "project-role",
            "namespace": "*",
            "access": [
              {"resource": "member", "action": "update", "effect": "allow"}
            ]
          }
        ]
      }
      ```

   2. As each `member-only` and `full` robot at both levels, add `robot-add-target` to A with this custom role using category 2's POST body and the actual custom role ID. Expect HTTP 201, verify the saved role as administrator, and remove the membership before the next attempt. Reset `robot-update-target` to Guest, then update it to the custom role as each robot; expect HTTP 200. As the same robot, change it from the custom role to Developer and then back to Guest; expect HTTP 200 for both changes and verify each saved role.
   3. As administrator, add `member-update-user` to A with only `Robot Test Member Updater`. Record this user's membership ID separately from `robot-update-target`'s ID. As `member-update-user`, change `robot-update-target` from Guest to the same custom role; expect HTTP 200 because the assigned permission is held by the caller. Verify the result as administrator and reset the target to Guest.
   4. As `member-update-user`, send `PUT /api/v2.0/projects/robot-permission-a/members/<own-member-id>` with `{"role_id":1}`. Expect HTTP 403. Verify as administrator that the caller still has the custom role. Repeat against `robot-update-target`'s membership ID and expect HTTP 403 with the target still Guest. These checks ensure restoring robot compatibility does not let a restricted human grant Project Admin to self or another user.

5. **Human custom-role permissions cannot be amplified through robot creation**

   1. As administrator, create a second custom role named `Robot Test Creator`. Reuse category 4's role creation body, changing its name and description and replacing `access` with exactly `robot:create`, `robot:read`, and `robot:list`, all with `effect: "allow"`. Add `robot-create-user` to A with this role only. Verify that this user has no repository or member permissions and no membership in B.
   2. As `robot-create-user`, send `POST /api/v2.0/robots` with the following body. Expect HTTP 201 because the requested `robot:read` grant is within the caller's own permissions. Record the created robot ID and verify its saved grants as administrator.

      ```json
      {
        "name": "creator-within-grant",
        "description": "Custom role robot creation control",
        "level": "project",
        "duration": 30,
        "disable": false,
        "permissions": [
          {
            "kind": "project",
            "namespace": "robot-permission-a",
            "access": [
              {"resource": "robot", "action": "read", "effect": "allow"}
            ]
          }
        ]
      }
      ```

   3. Repeat with unused names `creator-excess-pull` and `creator-excess-member`. Keep the valid body and append one extra permission per request: `{"resource":"repository","action":"pull","effect":"allow"}` for the first, and `{"resource":"member","action":"update","effect":"allow"}` for the second. Expect HTTP 403 for both. Using the administrator robot list, verify that neither robot exists and that the successful control robot's grants are unchanged.
   4. Submit each denied body as administrator with another unused name. Expect HTTP 201, proving the requests are valid and the human denials were authorization checks. Remove these control robots as administrator after verification.

6. **Missing permissions and project boundaries remain enforced**

   1. As each `member-reader` robot, confirm `GET /api/v2.0/projects/robot-permission-a/members` returns HTTP 200. Attempt to add the unassigned `robot-add-target` with Developer and to change the existing `robot-update-target` from Guest to Developer. Expect HTTP 403 for POST and PUT because `member:create` and `member:update` are absent. Verify as administrator that no membership was added and the target's role is unchanged.
   2. As each `member-only` and `full` robot at both levels, repeat the same POST and PUT against B, using B's membership ID for `robot-update-target`. Expect HTTP 403 for every request because all robot grants are scoped to A. Verify B's member list and Guest role are unchanged.
   3. Repeat an allowed member request in A with valid robot credentials but no Basic-auth header, then with an incorrect secret. Expect HTTP 401 and no membership changes. Repeat with the correct credentials and a reset fixture; expect the normal 201 or 200 response. Distinguish authentication failures from the permission regression.

7. **Upgrade verification, evidence, and cleanup**

   1. Before upgrading v2.15.3, retain the fixture robot credentials and record robot IDs, names, grants, project IDs, and member IDs. Remove `robot-add-target`'s memberships and reset `robot-update-target` to Guest in A and B. Upgrade the baseline instance to the v2.16.0 candidate using the supported upgrade procedure.
   2. Without recreating robots, changing their grants, or rotating their secrets, repeat categories 2, 3, and 6 with the existing credentials. Expect the same successful member operations and scope/missing-permission denials. Retrieve the robots as administrator and confirm their identities and explicit grants are preserved. Create the custom-role fixtures and run categories 4 and 5 on the upgraded instance as well.
   3. Compare the upgrade results with the fresh-install results. Record each case's build, installation path, robot level, actor, explicit grants, target role, HTTP status, error body, and administrator read-back. A passing custom-role operation does not compensate for a failed built-in-role operation; all three reported robot compatibility scenarios must pass while both human escalation attempts remain forbidden.
   4. Delete test robots, remove test memberships, delete the custom roles after they are unassigned, and remove the test projects and ordinary users. Delete local credential files created for this testcase. Retain the result matrix with secrets and authorization headers omitted.
