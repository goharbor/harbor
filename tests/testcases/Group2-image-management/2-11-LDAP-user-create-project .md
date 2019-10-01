Test 2-11 - User Create Project (LDAP Mode)
=======

# Purpose:

To verify that a non-admin user can create projects in (LDAP mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an LDAP or AD server. ( auth_mode is set to **ldap_auth** .) The user data is stored in an LDAP or AD server.
* A linux host with Docker CLI installed (Docker client).
* At least two non-admin users are in Harbor.

# Test Steps:

Same as Test 2-01 except that users are from LDAP/AD.

# Expected Outcome:
* Same as Test 2-01.

# Possible Problems:
None