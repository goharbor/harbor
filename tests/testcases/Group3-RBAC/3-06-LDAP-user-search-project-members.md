Test 3-06 - Search Project Members (LDAP Mode)
=======

# Purpose:

To verify that a non system admin user can search members of a project in LDAP mode.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an LDAP or AD server. ( auth_mode is set to **ldap_auth** .) The user data is stored in an LDAP or AD server.
* A linux host with Docker CLI installed (Docker client).
* At least five(5) non-admin users are in Harbor. 

# Test Steps:

Same as Test 3-03.


# Expected Outcome:
* Same as Test 3-03.

# Possible Problems:
None