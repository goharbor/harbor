Test 1-14 - LDAP Mode Admin Role General Functions
=======

# Purpose:

To verify that Harbor's UI of a user with system admin role works properly in LDAP mode.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an AD or LDAP server. (auth_mode is set to **ldap_auth** .)
* An Active Directory (AD) or LDAP server has been set up and it has a few users available for testing.

# Test Steps:

**NOTE:** The below non-admin user M should NOT be the same as the non-admin user in Test 1-07.

1. Assign an non-admin user M with system admin role and act as an admin user.
2. Repeat all steps in Test 1-07.

# Expected Outcome:

* A user with system admin role can perform all operations the same as the admin user.
* Same as Test 1-07.

# Possible Problems:
None