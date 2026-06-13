# Harbor PAT Implementation Integration Checklist

## Quick Start for Developers

This checklist guides the final steps to integrate PATs into Harbor and prepare for PR submission.

## Pre-Integration Verification

- [x] Database migration created: `make/migrations/postgresql/0200_2.17.0_schema.up.sql`
- [x] Go backend builds: `cd src && go build ./pkg/pat/...`
- [x] Security middleware integrated into chain
- [x] OpenAPI spec updated with 6 new endpoints + 5 schemas
- [x] CLI secret migration logic implemented
- [ ] **Next**: Generate API code and implement handlers

## Phase 1: Code Generation (15 min)

### 1.1 Generate API Code
```bash
cd /home/rossg/src/harbor/src
make generate_apis
```

**Expected Output**:
- New files in `server/v2.0/restapi/operations/user/`:
  - `create_personal_access_token.go`
  - `list_personal_access_tokens.go`
  - `get_personal_access_token.go`
  - `update_personal_access_token.go`
  - `delete_personal_access_token.go`
  - `refresh_personal_access_token_secret.go`

- New files in `server/v2.0/models/`:
  - `personal_access_token.go`
  - `personal_access_token_create_request.go`
  - `personal_access_token_created_response.go`
  - `personal_access_token_update_request.go`
  - `personal_access_token_refresh_request.go`

### 1.2 Verify Generation
```bash
ls -la src/server/v2.0/restapi/operations/user/personal_access_token*.go
ls -la src/server/v2.0/models/personal_access_token*.go
```

## Phase 2: Implement API Handler (30 min)

### 2.1 Update Handler File
Replace the template in `src/server/v2.0/handler/pat.go` with full implementation.

**Reference**: Copy patterns from `src/server/v2.0/handler/robot.go`

**Methods to Implement**:
```go
type patAPI struct {
    BaseAPI
    ctl pat.Controller
}

// Implement these operation handlers:
func (api *patAPI) CreatePersonalAccessToken(ctx, params) middleware.Responder
func (api *patAPI) ListPersonalAccessTokens(ctx, params) middleware.Responder
func (api *patAPI) GetPersonalAccessToken(ctx, params) middleware.Responder
func (api *patAPI) UpdatePersonalAccessToken(ctx, params) middleware.Responder
func (api *patAPI) DeletePersonalAccessToken(ctx, params) middleware.Responder
func (api *patAPI) RefreshPersonalAccessTokenSecret(ctx, params) middleware.Responder
```

**Authorization Pattern**:
```go
// Users can only access their own tokens
// System admins can access any user's tokens
currentUser, err := u.GetSecurityContext(ctx).User()
if currentUser.UserID != userID && !currentUser.SysAdminFlag {
    return u.SendError(ctx, errors.Forbidden())
}
```

### 2.2 Wire Handler
**File**: `src/server/v2.0/handler/handler.go`

Find the `Handlers` struct initialization and add:
```go
type Handlers struct {
    // ... existing handlers ...
    PersonalAccessTokenAPI: newPatAPI(),
}
```

### 2.3 Verify Build
```bash
cd /home/rossg/src/harbor/src
go build ./server/v2.0/handler/...
go test ./server/v2.0/handler/... -v
```

## Phase 3: Add Migration Call (5 min)

### 3.1 Update Migration Entry Point
**File**: `src/core/main.go`

Find where other data migrations are called (after orm.Context() is available):
```go
log.Info("Migrating OIDC CLI secrets to personal access tokens...")
if err := pat_migration.MigrateCliSecretsToLegacyPATs(orm.Context()); err != nil {
    log.Warningf("PAT migration failed (non-fatal): %v", err)
}
```

## Phase 4: Portal UI (60 min)

### 4.1 Create PAT Components
**Location**: `src/portal/src/app/base/account-settings/pat/`

**Create Files**:
```
pat-list.component.ts
pat-list.component.html
pat-list.component.scss
add-pat.component.ts
add-pat.component.html
add-pat.component.scss
pat-created.component.ts
pat-created.component.html
```

**Reference**: Model after `src/portal/src/app/base/project/robot-account/`

### 4.2 Add to Account Settings Modal
**File**: `src/portal/src/app/base/account-settings/account-settings-modal.component.html`

Add after CLI secret section:
```html
<pat-list *ngIf="!hidePatSection"></pat-list>
```

### 4.3 Add i18n Keys
**Files**: `src/portal/src/i18n/lang/*.json`

Add to `en-us-lang.json`:
```json
"PAT": {
    "TITLE": "Personal Access Tokens",
    "NEW_TOKEN": "Generate New Token",
    "DELETE_TOKEN": "Delete Token",
    "CREATE_SUCCESS": "Token created successfully",
    "DELETE_SUCCESS": "Token deleted successfully",
    "TOKEN_ONCE_WARNING": "This token will not be shown again. Copy it now.",
    "NAME": "Name",
    "DESCRIPTION": "Description",
    "EXPIRATION": "Expiration",
    "NEVER": "Never",
    "DAYS": "days",
    "LAST_USED": "Last Used",
    "EXPIRES": "Expires",
    "STATUS": "Status",
    "ACTIVE": "Active",
    "EXPIRED": "Expired",
    "DISABLED": "Disabled",
    "LEGACY": "Migrated CLI Secret"
}
```

Sync to other language files (at minimum en-us fallback).

### 4.4 Declare Components
**File**: `src/portal/src/app/base/base.module.ts`

Add to declarations:
```typescript
import { PatListComponent } from './account-settings/pat/pat-list.component';
import { AddPatComponent } from './account-settings/pat/add-pat.component';
import { PatCreatedComponent } from './account-settings/pat/pat-created.component';

@NgModule({
    declarations: [
        // ... existing ...
        PatListComponent,
        AddPatComponent,
        PatCreatedComponent,
    ]
})
```

## Phase 5: Testing (45 min)

### 5.1 Unit Tests
Create test files with ✅ baseline (can be extended):
- `src/pkg/pat/dao/dao_test.go` - Test DAO CRUD
- `src/pkg/pat/manager_test.go` - Test manager
- `src/controller/pat/controller_test.go` - Test controller (secret generation, refresh)
- `src/server/middleware/security/pat_test.go` - Test auth middleware
- `src/server/v2.0/handler/pat_test.go` - Test HTTP handlers
- `src/pkg/pat/migration/migrate_cli_secrets_test.go` - Test migration

Run:
```bash
cd /home/rossg/src/harbor/src
go test ./pkg/pat/... ./controller/pat/... ./server/middleware/security/... -v -race
```

### 5.2 Integration Tests
```bash
# Start Harbor locally (with migrations)
docker-compose up -d

# Verify migration ran
docker-compose logs core | grep "CLI secrets migration"

# Create a PAT
curl -X POST http://localhost:8080/api/v2.0/users/1/personal_access_tokens \
  -u admin:Harbor12345 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-token",
    "description": "Test token",
    "expires_in_days": 30
  }'

# Should return secret
# {
#   "id": 1,
#   "name": "test-token",
#   "secret": "hbr_pat_...",
#   "expires_at": 1234567890
# }

# Test docker login with PAT
docker login -u admin -p hbr_pat_... localhost:8080

# Test that legacy CLI secret still works (migrated)
docker login -u <user> -p <original-cli-secret> localhost:8080

# List user's PATs
curl http://localhost:8080/api/v2.0/users/1/personal_access_tokens \
  -u admin:Harbor12345

# Refresh PAT secret
curl -X PATCH http://localhost:8080/api/v2.0/users/1/personal_access_tokens/1 \
  -u admin:Harbor12345 \
  -H "Content-Type: application/json"

# Should return new secret

# Delete PAT
curl -X DELETE http://localhost:8080/api/v2.0/users/1/personal_access_tokens/1 \
  -u admin:Harbor12345
```

### 5.3 Portal UI Testing
1. Navigate to Account Settings
2. Find "Personal Access Tokens" section
3. Click "Generate New Token"
4. Enter name, description, expiration
5. Copy the displayed secret
6. Delete the token
7. Create another and try with `docker login`

## Phase 6: Code Review Readiness (15 min)

### 6.1 Code Quality
```bash
# Run linters
cd /home/rossg/src/harbor/src
make lint

# Check test coverage
go test ./pkg/pat/... ./controller/pat/... -cover

# Format code
go fmt ./pkg/pat/... ./controller/pat/...
```

### 6.2 Documentation
- [x] `PAT_IMPLEMENTATION_SUMMARY.md` - Architecture overview
- [x] `PAT_INTEGRATION_CHECKLIST.md` - This file
- [ ] Code comments for non-obvious logic
- [ ] Update Harbor docs with PAT usage guide

### 6.3 Commit Organization

Suggested commits:
1. `feat: add PAT database schema and migrations`
2. `feat: implement PAT model, DAO, manager, and controller`
3. `feat: add PAT security middleware for authentication`
4. `feat: add legacy CLI secret to PAT data migration`
5. `feat: add PAT REST API endpoints to Swagger spec`
6. `feat: implement PAT API handlers`
7. `feat: add PAT portal UI components`
8. `feat: add PAT localization strings`
9. `test: add unit tests for PAT system`

## Phase 7: PR Submission Checklist

### Before Creating PR:
- [ ] All phases complete
- [ ] All tests pass: `go test ./... -race`
- [ ] Code formatted and linted
- [ ] No breaking changes to existing APIs
- [ ] Swagger spec validates: `swagger validate api/v2.0/swagger.yaml`
- [ ] Portal builds: `npm run build -- --configuration production`
- [ ] Migration tested with real data
- [ ] Documentation updated

### PR Title
```
feat: implement Personal Access Tokens (PAT) system for Harbor

- Works with all auth modes (OIDC, DB, LDAP)
- Replaces OIDC CLI secret with enterprise-grade token management
- Automatic migration of legacy CLI secrets
- Includes web UI and API for token lifecycle management
```

### PR Description
Include:
- Summary of changes
- Security properties
- Backward compatibility notes
- Migration details
- Testing checklist
- Screenshots of UI (if applicable)

## Troubleshooting

### Swagger Generation Fails
- Ensure `api/v2.0/swagger.yaml` is valid YAML
- Check for duplicate operation IDs
- Verify all schema `$ref` paths exist in definitions

### Handler Compilation Errors
- Check that generated operation types are imported
- Verify method signatures match generated interfaces
- Use robot handler as reference for pattern

### PAT Authentication Fails
- Check PAT middleware is in security chain before oidcCli
- Verify `hbr_pat_` prefix is correct
- Check PBKDF2 verification logic matches CreateSec salt/secret

### Portal Component Issues
- Verify ng-swagger-gen service is imported
- Check i18n keys are defined in all language files
- Ensure components are declared in base.module.ts

## Success Criteria

When implementation is complete, you should be able to:

✅ Create PAT via API with 30-day expiry  
✅ List PATs for logged-in user  
✅ `docker login` with `hbr_pat_XXXXX` password  
✅ Delete PAT and lose access immediately  
✅ Refresh PAT secret and get new token  
✅ Existing CLI secrets work after migration  
✅ Portal UI shows PAT management section  
✅ Tokens marked as legacy show in UI  
✅ All unit tests pass  
✅ Zero breaking changes to existing features  

## Timeline Estimate

| Phase | Est. Time | Status |
|-------|-----------|--------|
| 1. Code Generation | 15 min | ⏳ |
| 2. API Handler | 30 min | ⏳ |
| 3. Migration Call | 5 min | ⏳ |
| 4. Portal UI | 60 min | ⏳ |
| 5. Testing | 45 min | ⏳ |
| 6. Code Review Ready | 15 min | ⏳ |
| **Total** | **~3 hours** | ⏳ |

**Total Backend Work (Completed)**: ~2 hours ✅

## Next Immediate Action

```bash
cd /home/rossg/src/harbor/src
make generate_apis
```

Then proceed with Phase 2 (Handler Implementation).
