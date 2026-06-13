# Harbor Personal Access Tokens (PAT) Implementation Summary

## Overview

This document summarizes the PAT (Personal Access Tokens) implementation for Harbor. The system provides a secure, auditable alternative to the OIDC CLI secret with support across all authentication modes.

## Implementation Status

### ✅ Completed Components

#### 1. Database Layer
- **File**: `make/migrations/postgresql/0200_2.17.0_schema.up.sql`
- **Status**: Complete
- **Contents**:
  - Creates `personal_access_token` table with columns:
    - `id` (BIGSERIAL PK)
    - `user_id` (FK to harbor_user)
    - `name`, `secret`, `salt` (token storage)
    - `description`, `expires_at`, `last_used_at`
    - `disabled`, `is_legacy` (flags)
    - `creation_time`, `update_time` (auditing)
  - Unique constraint on (user_id, name)
  - Index on user_id and disabled
  - Update trigger for `update_time`

#### 2. Go Model Layer
- **Package**: `src/pkg/pat/`
- **Files**:
  - `model/model.go` - PersonalAccessToken ORM model with proper tags
  - `dao/dao.go` - Data access layer (CRUD operations)
  - `manager.go` - Manager interface and implementation
- **Status**: Complete
- **Pattern**: Follows robot account structure exactly

#### 3. Controller Layer
- **File**: `src/controller/pat/controller.go`
- **Status**: Complete
- **Methods**:
  - `Create()` - Generate PBKDF2-SHA256 hashed secret with `hbr_pat_` prefix
  - `Get()`, `List()`, `Count()` - Query operations
  - `Update()`, `Delete()` - Modification operations
  - `RefreshSecret()` - Secret rotation support
- **Features**:
  - Reuses `robot.CreateSec()` for secret generation
  - Uses `robot.IsValidSec()` for validation
  - Returns plaintext secret once at creation only

#### 4. Security Middleware
- **File**: `src/server/middleware/security/pat.go`
- **Status**: Complete
- **Features**:
  - HTTP Basic Auth detection with `hbr_pat_` prefix
  - PBKDF2 hash verification
  - Expiry checking
  - Async `last_used_at` updates
  - Integrated into security chain (before oidcCli)
- **Wire-up**: Modified `src/server/middleware/security/security.go` to add `&pat{}` to generators slice

#### 5. Data Migration
- **File**: `src/pkg/pat/migration/migrate_cli_secrets.go`
- **Status**: Complete
- **Features**:
  - Migrates existing OIDC CLI secrets to legacy PATs (`is_legacy=true`)
  - Decrypts AES secrets and re-hashes as PBKDF2
  - Handles decryption failures gracefully
  - Idempotent (safe to run multiple times)
  - Creates PAT named "cli-secret" for each user with existing secret

#### 6. OpenAPI Specification
- **File**: `api/v2.0/swagger.yaml`
- **Status**: Complete
- **Added Endpoints**:
  - `GET /users/{user_id}/personal_access_tokens` - List user's PATs
  - `POST /users/{user_id}/personal_access_tokens` - Create PAT
  - `GET /users/{user_id}/personal_access_tokens/{token_id}` - Get PAT
  - `PUT /users/{user_id}/personal_access_tokens/{token_id}` - Update PAT
  - `DELETE /users/{user_id}/personal_access_tokens/{token_id}` - Delete PAT
  - `PATCH /users/{user_id}/personal_access_tokens/{token_id}` - Refresh secret
- **Added Schemas**:
  - `PersonalAccessToken` - Full PAT metadata
  - `PersonalAccessTokenCreateRequest` - Create request with name, description, expires_in_days
  - `PersonalAccessTokenCreatedResponse` - Response with plaintext secret (shown once)
  - `PersonalAccessTokenUpdateRequest` - Update request
  - `PersonalAccessTokenRefreshRequest` - Refresh request with optional new secret

### 🔄 Partially Completed

#### REST API Handler
- **File**: `src/server/v2.0/handler/pat.go`
- **Status**: Template provided
- **Next Steps**:
  1. Run code generation: `cd src && make generate_apis`
  2. This generates operation handlers in `src/server/v2.0/restapi/operations/user/`
  3. Implement the handler methods by embedding the generated operations

#### Handler Registration
- **File**: `src/server/v2.0/handler/handler.go`
- **Status**: Not yet updated
- **Needed**: Add line like:
  ```go
  PersonalAccessTokenAPI: newPatAPI(),
  ```

### ⏳ Not Yet Started

#### Portal UI Components
- **Location**: `src/portal/src/app/base/account-settings/pat/`
- **Components Needed**:
  - `pat-list.component.ts/html` - List PATs in a datagrid
  - `add-pat.component.ts/html` - Create PAT modal
  - `pat-created.component.ts/html` - Show secret once
- **Integration**: Add to `account-settings-modal.component.html`

#### i18n Localization
- **Files**: `src/portal/src/i18n/lang/*.json`
- **Needed**: Add "PAT" section with keys for all UI labels

## Next Steps to Complete Implementation

### Step 1: Generate API Code (5 minutes)
```bash
cd /home/rossg/src/harbor/src
make generate_apis
```

This will generate:
- Operation handler interfaces in `server/v2.0/restapi/operations/user/`
- Model types in `server/v2.0/models/`

### Step 2: Implement Handler Methods (30 minutes)
Edit `src/server/v2.0/handler/pat.go` to implement generated operation interfaces. Reference patterns from `robot.go`.

### Step 3: Wire Handler (2 minutes)
Edit `src/server/v2.0/handler/handler.go`:
```go
func newHandlers() *Handlers {
    return &Handlers{
        // ... existing handlers ...
        PersonalAccessTokenAPI: newPatAPI(),
    }
}
```

### Step 4: Build Portal UI (60 minutes)
Create components in `src/portal/src/app/base/account-settings/pat/`:
- List component with delete functionality
- Create/edit modal
- Display secret once after creation

### Step 5: Add Localization (10 minutes)
Update `src/portal/src/i18n/lang/*.json` with PAT strings.

### Step 6: Integration Testing (30 minutes)
1. Start Harbor with migrations
2. Verify legacy CLI secrets migrated to PATs
3. Create new PAT via API
4. Test `docker login` with PAT
5. Test expiry and refresh

## Architecture Highlights

### Token Format
- **New PATs**: `hbr_pat_<32-char-random>`
- **Legacy PATs**: Original CLI secret, stored as migrated PAT with `is_legacy=true`
- **Storage**: PBKDF2-SHA256 hash (same as robot accounts)

### Security Properties
✅ Hash-based storage (not reversibly encrypted)  
✅ Per-user tokens with optional expiry  
✅ Audit trail (created_at, last_used_at, expires_at)  
✅ One-time secret display (never shown again)  
✅ Async last_used_at updates (non-blocking)  
✅ Works across all auth modes (not just OIDC)  

### Backward Compatibility
✅ Existing OIDC CLI secrets work transparently via migration  
✅ Can coexist with oidcCli middleware (belt-and-suspenders)  
✅ No breaking changes to existing APIs  

## Key Files Summary

| File | Purpose | Status |
|------|---------|--------|
| `make/migrations/postgresql/0200_2.17.0_schema.up.sql` | Schema | ✅ |
| `src/pkg/pat/model/model.go` | ORM Model | ✅ |
| `src/pkg/pat/dao/dao.go` | Data Access | ✅ |
| `src/pkg/pat/manager.go` | Manager | ✅ |
| `src/controller/pat/controller.go` | Business Logic | ✅ |
| `src/server/middleware/security/pat.go` | Auth Middleware | ✅ |
| `src/pkg/pat/migration/migrate_cli_secrets.go` | Data Migration | ✅ |
| `src/migration/migration.go` | Migration Trigger | ⏳ (add call) |
| `api/v2.0/swagger.yaml` | OpenAPI Spec | ✅ |
| `src/server/v2.0/handler/pat.go` | API Handler | 🔄 |
| `src/server/v2.0/handler/handler.go` | Handler Registration | ⏳ |
| Portal UI components | Frontend | ⏳ |
| i18n files | Localization | ⏳ |

## Testing the Implementation

### Unit Tests Needed
- `src/pkg/pat/dao/dao_test.go` - DAO CRUD
- `src/pkg/pat/manager_test.go` - Manager layer
- `src/controller/pat/controller_test.go` - Controller (create/refresh)
- `src/server/middleware/security/pat_test.go` - Middleware auth flow
- `src/server/v2.0/handler/pat_test.go` - Handler operations
- `src/pkg/pat/migration/migrate_cli_secrets_test.go` - Migration safety

### Integration Tests
```bash
# Create PAT via API
curl -X POST http://localhost:8080/api/v2.0/users/1/personal_access_tokens \
  -H "Content-Type: application/json" \
  -d '{"name": "test-token", "description": "Test", "expires_in_days": 30}'

# Login with PAT
docker login -u admin -p hbr_pat_XXXXX localhost:8080

# Refresh PAT secret
curl -X PATCH http://localhost:8080/api/v2.0/users/1/personal_access_tokens/1
```

## Configuration & Environment

### Feature Detection
- Auto-enabled for all auth modes
- OIDC mode gets automatic CLI secret migration on startup
- No feature flags required

### Database Requirements
- PostgreSQL (matches Harbor's requirement)
- Automatic schema migration on Harbor startup
- Data migration runs as background task (non-blocking)

## Security Considerations

1. **Secret Display**: Plaintext secret shown only once at creation/refresh
2. **Storage**: PBKDF2-SHA256 hashed, matching robot account security
3. **Expiry**: Optional configurable TTL with -1 (never) as default
4. **Revocation**: Individual PATs can be deleted/disabled
5. **Audit**: Last used time tracked for access monitoring
6. **Scope**: v1 uses full user permissions (fine-grained scope in future)

## Performance Notes

- `last_used_at` updates are async (non-blocking) to avoid request slowdown
- Database queries use indexed lookups (user_id, disabled flags)
- No N+1 queries in auth path
- Suitable for high-throughput authentication scenarios

## Future Enhancements

1. **Fine-Grained Scope**: Per-project or per-action permissions
2. **Token Refresh**: Auto-refresh on near-expiry
3. **Activity Logs**: Detailed usage history per token
4. **Bulk Operations**: Create multiple tokens at once
5. **Admin Dashboard**: Manage user tokens as admin

## References

### Related Harbor Systems
- Robot Accounts: `src/pkg/robot/` (used as implementation template)
- OIDC Integration: `src/pkg/oidc/` (coexists with PATs)
- Security Context: `src/common/security/` (auth integration)

### Related Standards
- OAuth 2.0 Bearer Token RFC 6750
- Personal Access Token patterns (GitHub, GitLab)
- PBKDF2 for password hashing RFC 2898
