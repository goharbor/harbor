Test 4-05 - User with Admin Role Views Logs
=======

# Purpose:

To verify that a user with system admin role can views logs.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is configured with LDAP/AD authentication. Users are stored in an external LDAP or AD directory.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

**NOTE:** The below non-admin user A should NOT be the same as the non-admin user in Test 4-03.

1. Assign an non-admin user M with system admin role and act as an admin user. 
2. Repeat all steps in Test 4-03.


# Expected Outcome:

* A user with system admin role can perform all operations the same as the admin user. 
* Outcome should be the same as Test 4-03.

# Possible Problems:
None