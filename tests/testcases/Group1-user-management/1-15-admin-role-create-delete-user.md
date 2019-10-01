Test 1-15 - Admin Role User Create, Delete and Recreate a User(DB Mode)
=======

# Purpose:

To verify that an admin user can create/delete/recreate a user when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

**NOTE:** The below non-admin user M should NOT be the same as the non-admin user in Test 1-09.

1. Assign an non-admin user M with system admin role and act as an admin user.
2. Repeat all steps in Test 1-09.

# Expected Outcome:

* A user with system admin role can perform all operations the same as the admin user.
* Same as Test 1-09.

# Possible Problems:
None