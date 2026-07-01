# OAuth2/OIDC Bearer Token API Auth — Future Enhancement

**Status: Shelved.** This document describes a planned future enhancement to allow OAuth2 bearer tokens from an external OIDC provider to authenticate API requests independent of Harbor's primary authentication mode (DB, LDAP, etc.). This is not currently implemented.

## Problem Statement

Harbor already has the plumbing to validate `Authorization: Bearer <JWT>` tokens against an external OIDC provider's live JWKS on every `/api` or `/service/token` request via `src/server/middleware/security/idtoken.go`. However, this is hard-gated to only work when OIDC is the instance's *primary* authentication mode (`AuthMode == common.OIDCAuth`).

Use case: An instance using DB auth for the web UI should still be able to let CI pipelines/scripts present an external IdP's OAuth2 access token (e.g. from Keycloak) on API calls as an alternative to basic auth, robot accounts, or PATs.

Current situation: OIDC configuration (endpoint, client ID, etc.) is stored independently of `AuthMode`, so the underlying validation infrastructure exists — it just needs to be unlocked via a gating condition change.

## Design Overview

### Core Changes
1. **New boolean config flag**: `oidc_api_token_auth` (default: false)
   - Independent of `AuthMode`
   - Stored alongside existing OIDC settings (endpoint, client ID, etc.)
   - Exposed in the UI via Configuration → Authentication page

2. **Middleware gate change**: Replace the `AuthMode == OIDCAuth` check in `idtoken.go:Generate()` with a config lookup
   - If flag is enabled and a valid Bearer token is present, validate it against the configured OIDC provider's JWKS
   - User lookup via `user.Ctl.GetBySubIss(subject, issuer)`
   - No new account auto-provisioning (that stays in the login callback only)

3. **No authentication logic changes**: The rest of `idToken.Generate()` is unchanged
   - JWKS verification via `oidc.VerifyToken()` (existing)
   - User lookup by subject + issuer (existing)
   - Group injection from claims (existing)

## Implementation Path

### Backend Config (4 files)

**1. `src/common/const.go`**
```go
OIDCAPITokenAuth = "oidc_api_token_auth"
```

**2. `src/lib/config/metadata/metadatalist.go`**
Add metadata entry (next to `OIDCAutoOnboard`):
```go
{
    Name: common.OIDCAPITokenAuth,
    Scope: UserScope,
    Group: OIDCGroup,
    DefaultValue: "false",
    ItemType: &BoolType{},
    Description: "Allow OAuth2/OIDC bearer access tokens to authenticate API requests independent of the primary auth mode",
}
```

**3. `src/lib/config/models/model.go`**
Add field to `OIDCSetting` struct:
```go
APITokenAuth bool `json:"oidc_api_token_auth"`
```

**4. `src/lib/config/userconfig.go`**
Wire into `OIDCSetting()` function:
```go
APITokenAuth: mgr.Get(ctx, common.OIDCAPITokenAuth).GetBool(),
```

### REST API (Swagger)

**`api/v2.0/swagger.yaml`**
- Add `oidc_api_token_auth` to `Configurations` definition (PUT body)
- Add `oidc_api_token_auth` to `ConfigurationsResponse` definition (GET response)
- Mirror existing `oidc_auto_onboard` entries exactly (same `BoolConfigItem` ref pattern)
- Run `make swagger` to regenerate Go models (no manual handler code needed — fully metadata-driven)

### Middleware (2 files)

**`src/server/middleware/security/idtoken.go`**
Replace the `AuthMode` check:
```go
// OLD
if lib.GetAuthMode(ctx) != common.OIDCAuth {
    return nil
}

// NEW
setting, err := config.OIDCSetting(ctx)
if err != nil || !setting.APITokenAuth {
    return nil
}
```
- Fetch config once at the top (currently fetched later)
- Reuse `setting` for the `UserInfoFromIDToken()` call below (avoid double-fetch)
- Keep path-prefix check (`/api` or `/service/token`)
- Keep all validation logic unchanged (JWKS, user lookup, group injection)

**`src/server/middleware/security/idtoken_test.go`**
- Rewrite `TestIDToken` to test config flag instead of `AuthMode`
- Add test case: flag enabled + valid Bearer token → context generated
- Mock `user.Ctl.GetBySubIss` per existing patterns in `oidc_cli_test.go`

### Frontend UI (3 files)

**`src/portal/src/app/base/left-side-nav/config/config.ts`**
Add field to `Configuration` class:
```ts
oidc_api_token_auth?: BoolValueItem;
```
Initialize in constructor:
```ts
this.oidc_api_token_auth = new BoolValueItem(false, true);
```

**`src/portal/src/app/base/left-side-nav/config/auth/config-auth.component.html`**
Add checkbox block after existing `oidcAutoOnboard` (lines ~933-961):
```html
<clr-checkbox-container>
    <label for="oidcApiTokenAuth">
        {{ 'CONFIG.OIDC.OIDC_API_TOKEN_AUTH' | translate }}
        <clr-tooltip>
            <clr-icon clrTooltipTrigger shape="info-circle" size="24"></clr-icon>
            <clr-tooltip-content clrPosition="top-right" clrSize="lg" *clrIfOpen>
                <span>{{ 'TOOLTIP.OIDC_API_TOKEN_AUTH' | translate }}</span>
            </clr-tooltip-content>
        </clr-tooltip>
    </label>
    <clr-checkbox-wrapper>
        <input type="checkbox" clrCheckbox name="oidcApiTokenAuth" id="oidcApiTokenAuth"
            [disabled]="disabled(currentConfig.oidc_api_token_auth)"
            [(ngModel)]="currentConfig.oidc_api_token_auth.value" />
    </clr-checkbox-wrapper>
</clr-checkbox-container>
```

**`src/portal/src/app/base/left-side-nav/config/auth/config-auth.component.ts`**
- No changes needed — `getChanges()` already includes `oidc_`-prefixed fields via prefix filter
- No dependent-field logic needed (unlike `changeAutoOnBoard()`)

### Internationalization

**`src/portal/src/i18n/lang/en-us-lang.json`**
Add two keys (follow existing `OIDC_AUTOONBOARD` pattern):

Under `CONFIG.OIDC`:
```json
"OIDC_API_TOKEN_AUTH": "Enable OAuth2 API token auth"
```

Under `TOOLTIP`:
```json
"OIDC_API_TOKEN_AUTH": "Allow bearer tokens from the configured OIDC provider to authenticate API requests without using the primary authentication mode (e.g., use OIDC tokens for CI/API access while keeping DB or LDAP for web UI login)"
```

Other locale files (`de-de-lang.json`, `fr-fr-lang.json`, etc.) can be deferred to translators per project convention.

## Testing Strategy

1. **Unit tests**: `go test ./server/middleware/security/...`
   - Flag disabled → Generator returns nil
   - Flag enabled, no Bearer token → Generator returns nil
   - Flag enabled, valid Bearer token → Generator returns populated security context
   - All tests pass regardless of `AuthMode` setting

2. **Integration test**: Manual setup
   - Set `AuthMode = db_auth` (web UI uses DB credentials)
   - Enable the new toggle via Configuration page
   - Configure a real OIDC provider (Keycloak/Dex with test user)
   - Issue an API call with `Authorization: Bearer <valid-token>`
   - Confirm request authenticates as the linked Harbor user

3. **UI test**: Manual
   - Toggle the checkbox on Configuration → Authentication page
   - Confirm it persists across page reload
   - Confirm existing "Automatic onboarding" checkbox still works independently

4. **Build test**: `go build ./...` && `make lint` && `make swagger`

## Scope & Non-Scope

### In Scope
- Gating condition only (config flag → enable/disable)
- Reuses all existing OIDC infrastructure (JWKS validation, user lookup, group injection)
- Minimal code changes (mostly config plumbing)

### Out of Scope
- New account auto-provisioning (stays in login callback, not per-request middleware)
- Email-based user matching (stays for interactive login only)
- Opaque/non-JWT token introspection (would need RFC 7662 `/introspect` endpoint support — not planned here)
- Token scope/permission validation (inherits user's full permissions, like OIDC CLI secret today)

## Risk Assessment

**Low risk**: Entirely gated by default (flag off), reuses proven JWKS validation path, no new account creation logic.

**Potential issues**:
- Configuration overshoot (admins forget they enabled it) → mitigated by explicit UI checkbox
- User confusion (multiple auth paths available) → mitigated by clear UI labels and tooltip

## References

- Current OIDC Bearer token handling: `src/server/middleware/security/idtoken.go`
- OIDC token verification: `src/pkg/oidc/helper.go:223` (`VerifyToken`)
- Config pattern: `src/lib/config/metadata/metadatalist.go` (model after `OIDCAutoOnboard`)
- Frontend pattern: `src/portal/src/app/base/left-side-nav/config/auth/config-auth.component.ts` (model after existing OIDC checkbox)
