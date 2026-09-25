/* track whether sysadmin_flag was set by an operator or synced from the IdP admin group,
   so that LDAP/OIDC group-based admin sync never silently overrides an explicit manual decision */
ALTER TABLE harbor_user ADD COLUMN IF NOT EXISTS sysadmin_flag_source varchar(16) DEFAULT NULL;

/* preserve existing admins as-is, so this migration can never auto-revoke one on next login.
   Existing non-admins are deliberately left with a NULL (sync-eligible) source instead, even
   though a few may have been manually demoted in the past and could get re-synced to admin:
   that's accepted, since IsSysAdmin() already ORs in the live LDAP/OIDC role, so a stale false
   flag never actually revoked their real access anyway. Only future manual changes are immune
   to sync, via the 'manual' tag SetSysAdminFlag now always sets. */
UPDATE harbor_user SET sysadmin_flag_source = 'manual' WHERE sysadmin_flag = true AND sysadmin_flag_source IS NULL;
