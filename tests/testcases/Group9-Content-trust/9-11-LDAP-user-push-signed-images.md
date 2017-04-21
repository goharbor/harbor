Test 9-11 User push signed images(LDAP mode)
=======

# Purpose:

To verify user can sign and push images(LDAP mode)

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.  
* Harbor is set to authenticate against an LDAP or AD server.(auth_mode is set to ldap_auth.) The user data is stored in an LDAP or AD server.  
* A Linux host with Docker CLI(Docker client) installed.

# Test Steps:

Same as Test 9-01 except that users are from LDAP/AD.

# Expected Outcome:

* Same as Test 9-01.

# Possible Problems:
None
