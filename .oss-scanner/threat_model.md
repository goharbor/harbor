# Threat model

## What this project does and where untrusted input enters
Harbor is a cloud native OCI registry. It stores and distributes container images, Helm charts and other OCI artifacts, and adds projects, RBAC, robot accounts, replication, proxy cache, vulnerability scanning, signing/SBOM accessories, webhooks, quotas, retention and immutability on top of the embedded distribution registry.

Go services (all under `src/`):
- `core` (`src/core/`): REST API v2.0 (`src/server/v2.0/handler/`, generated from `api/v2.0/swagger.yaml`), the OCI Distribution API front (`src/server/registry/`) and the middleware chain in front of the registry (`src/server/middleware/`), the token service (`src/core/service/token/`), and authentication.
- `jobservice` (`src/jobservice/`): async jobs (replication, GC, scan, retention, webhooks), driven through Redis.
- `registryctl` (`src/registryctl/`): internal control API next to the registry.
- `exporter` (`src/cmd/exporter/`) and the standalone DB migrator (`src/cmd/standalone-db-migrator/`).
- The Angular portal (`src/portal/src/`).

Untrusted input:
- Every HTTP request to core from anonymous users, authenticated users of any role, and robot accounts: REST API, `/v2/` registry API, `/service/token`, `/c/` login and OIDC callback.
- Artifact content pushed by users: manifests, indexes, configs, layers, and artifact-type specific processing (`src/controller/artifact/processor/`), accessories such as signatures and SBOMs (`src/pkg/accessory/`).
- Responses from remote registries during replication (`src/pkg/reg/adapter/`) and proxy cache (`src/controller/proxy/`, `src/server/middleware/repoproxy/`). A remote registry may be malicious.
- Responses from external scanners (`src/pkg/scan/rest/`), OIDC providers (`src/pkg/oidc/`, `src/core/auth/oidc/`), LDAP (`src/core/auth/ldap/`), UAA and auth proxy (`src/core/auth/uaa/`, `src/core/auth/authproxy/`).
- Webhook target URLs configured by project admins (`src/pkg/notifier/`, `src/controller/event/handler/webhook/`), and replication/proxy endpoint URLs configured by admins.
- Query parameters that become SQL (`src/lib/q/`, `src/lib/orm/`).

Trusted: system administrator configuration (`harbor.yml`, environment), the internal network between Harbor components, internal shared secrets (`src/common/security/secret/`), and PostgreSQL and Redis/Valkey as services. Values stored in the database or Redis keep the trust level of where they came from: data derived from users, artifacts or remote services stays untrusted after Harbor persists it.

## Components that matter most / least
Most:
- Authentication and authorization: security context resolution (`src/server/middleware/security/`, `src/common/security/`), RBAC/permission evaluation (`src/pkg/permission/`), robot accounts (`src/pkg/robot/`, `src/controller/robot/`), registry token issuing and scope checks (`src/core/service/token/`, `src/server/middleware/v2auth/`). Cross-project access, privilege escalation between roles, and robot accounts exceeding their granted scope are the core of Harbor's security.
- Tenant/project isolation in every API handler: a user with a role in project A must not read or modify project B. IDOR in `src/server/v2.0/handler/` is high impact.
- Proxy cache and replication: SSRF via endpoint URLs or redirects, credential leakage to remote registries, cache poisoning between projects.
- Webhooks: SSRF to internal services, leaking secrets in payloads.
- Secrets at rest and in responses: registry credentials, robot secrets, OIDC client secrets and CLI secrets must not reach anyone not entitled to them, through API responses, logs or webhook payloads. Returning a robot secret to the caller who creates or refreshes that robot (`src/server/v2.0/handler/robot.go`) is intended.
- SQL built from user filters (`src/lib/q/`, `src/lib/orm/`).
- Portal XSS from artifact metadata, labels, repository descriptions and scan results.

Least / out of scope:
- `tests/`, `make/` (packaging, installer and photon images), `tools/`, `.github/`.
- The embedded upstream registry (distribution/distribution) itself, unless Harbor's integration with it is at fault.
- Trivy and other scanner adapters (separate repositories).
- Generated code in `src/server/v2.0/models` and `src/server/v2.0/restapi` (report the swagger spec or template instead).
- Issues that require system administrator privileges, unless they escape to the host or leak secrets that admins cannot otherwise read.

## How to exercise it
- Binaries are built: `harbor-core`, `harbor-jobservice`, `harbor-registryctl`, `harbor-exporter`, `harbor-db-migrator` in `/usr/local/bin`. A full running Harbor stack is not provisioned in this image.
- Generated API server code is present in `src/server/v2.0`; `swagger` is installed.
- Pure unit tests: `cd /src/src && go test ./pkg/permission/...` (no services needed). Many other packages, including `server/middleware/v2auth`, `server/middleware/security` and `server/middleware/csrf`, connect to PostgreSQL in `TestMain` and need `harbor-test-services` first.
- DB/Redis tests: run `harbor-test-services` first (starts PostgreSQL 18 on localhost:5432 and Redis on localhost:6379; env vars are preset), then e.g. `cd /src/src && go test -p 1 ./pkg/robot/... ./controller/...`. Packages needing LDAP, SMTP or a full stack (`core/api`, `core/controllers`, `*/ldap`, `common/utils/email`) will fail and are not a signal.
- Portal: `cd /src/src/portal`, dependencies installed with `npm ci` and a production build in `dist/`. Karma unit tests run headless with the installed Chromium: `npm run test:headless`.

## How you rate severity
The trust boundary is between a Harbor user (anonymous, project guest, developer, maintainer, project admin, robot account) and anything that user is not granted.
- **critical**: unauthenticated access to private artifacts or APIs, authentication bypass, privilege escalation to system admin, RCE in any Harbor service, SQL injection reachable by a non-admin.
- **high**: cross-project read or write by an authenticated user, robot account exceeding its permissions, SSRF reaching internal services from a non-admin, leaking credentials or secrets (robot secrets, registry credentials, OIDC secrets) to a user not entitled to them, stored XSS in the portal reachable by a non-admin.
- **medium**: SSRF or secret exposure requiring project admin, cache poisoning between projects, information disclosure of metadata (repository names, tags, user lists) across projects, DoS via a single unauthenticated or low-privilege request (resource exhaustion, unbounded allocation).
- **low**: issues requiring system admin, crashes recovered by restart, missing hardening without a demonstrated impact.

Reports should include the request sequence (curl or Go test) that triggers the issue, the role used, and the commit.

## Anything to leave alone
- Default credentials in test fixtures and CI env (`Harbor12345`, `root123`, `tempString`).
- Findings in dependencies with no reachable path from Harbor code (report upstream).
- Rate limiting or brute-force protection absent on login; Harbor delegates that to the deployment.
