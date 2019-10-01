Test 4-04 - User with Admin Role Views Logs (DB Mode)
=======

# Purpose:

To verify that a user with system admin role can views logs when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

**NOTE:** The below non-admin user A should NOT be the same as the non-admin user in Test 4-03.

1. Assign an non-admin user A with system admin role.
2. Repeat all steps in Test 4-03.


# Expected Outcome:

* A user with admin role can perform all operations the same as the admin user.
* Outcome should be the same as Test 4-03.

# Possible Problems:
None