Test 3-22 - Admin Role User Search Project Members (DB Mode)
=======

# Purpose:

To verify that a user with system admin role can search members of a project when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least five(5) non-admin users are in Harbor.
* At least 5 members in a project that admin user is not a member of.


# Test Steps:

**NOTE:** The below non-admin user M should NOT be the same as the non-admin user in Test 3-12.

1. Assign an non-admin user M with system admin role and act as an admin user.
2. Repeat all steps in Test 3-12.

# Expected Outcome:

* A user with system admin role can perform all operations the same as the admin user.
* Same as Test 3-12.

# Possible Problems:
None