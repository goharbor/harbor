# AI Agent Working Notes — Harbor (golder fork)

## Branch Strategy

This fork maintains two long-lived branches for the PAT (Personal Access Tokens) feature work:

| Branch | Base | Purpose |
|--------|------|---------|
| `feature/personal-access-tokens` | `upstream/main` | Clean upstream PR — rebase on `upstream/main` before submitting |
| `develop` | `upstream/main` + local patches | Internal testing — carries schema changes, OIDC fixes, and CI customizations our test hosts depend on |

### Rules for AI Agents

- **Never rebase `develop`** onto `upstream/main` using `git rebase` — it contains intentional merge commits and local CI patches that must stay.
- **Always rebase `feature/personal-access-tokens`** on `upstream/main` before raising a PR.
- **Feature work belongs on both branches**: commit on `feature/personal-access-tokens` first, then cherry-pick onto `develop` (or vice-versa if testing on `develop` first).
- **`develop` divergence is intentional** — it carries patches not intended for upstream (ghcr.io CI, OIDC non-fatal fixes, schema extensions, etc.).
- **Do not collapse the branches** — they serve different purposes and must remain separate.

## Schema Migration Numbering

Latest migration on `upstream/main`: `0190_2.16.0_schema.up.sql`  
PAT migration (on both branches): `0200_2.17.0_schema.up.sql`

Do not renumber the PAT migration unless upstream merges a conflicting `0200` first.

## Working Notes (PAT_*.md)

The `PAT_*.md` documents at the repo root are **working notes only** — do not commit them. They are generated during development and should be deleted when work is complete:

- `PAT_BUILD_VERIFICATION.md` — build status snapshot
- `PAT_DELIVERY_SUMMARY.md` — delivery tracking
- `PAT_FILES_MANIFEST.md` — file listing
- `PAT_IMPLEMENTATION_SUMMARY.md` — implementation overview
- `PAT_INTEGRATION_CHECKLIST.md` — integration checklist
- `PAT_TESTS_SUMMARY.md` — test summary
- `PAT_UNIT_TESTS_COMPLETION.md` — test completion report

## Current Status

- ✅ PAT implementation committed on `develop` (commit aa62fc221)
- ⏳ Feature branch to be created from `upstream/main` with cherry-picked commits
- ⏳ `AGENTS.md` to be committed on both branches
