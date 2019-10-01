Test 3-08 - Define Harbor Admin By DN (LDAP Mode)
=======

# Purpose:

To verify that harbor admin can be defined by LDAP group DN

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an LDAP or AD server. ( auth_mode is set to **ldap_auth** .) The user data is stored in an LDAP or AD server.
* A linux host with Docker CLI installed (Docker client).
* At least five(5) non-admin users are in Harbor.

# Test Steps:


1. Create group harbor_sys_admin in LDAP.
1. Create LDAP user sys_admin in LDAP, and add sys_admin to the member of group harbor_sys_admin.
1. Login as admin user, Go to Administration -> Configuration -> Authentication,  set the configuration: "LDAP Groups With Admin Privilege" with the DN of harbor_sys_admin.
1. Login in with LDAP user sys_admin.


# Expected Outcome:

1. The sys_admin user can have admin privileges, such as change the configuration, add/remove replication policy, manage repositories etc.

# Possible Problems:
None