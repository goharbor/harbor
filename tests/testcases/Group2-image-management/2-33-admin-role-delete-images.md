Test 2-33 - Admin Role User Delete Images
=======

# Purpose:

To verify that a user with system admin role can delete images owned by other users using local database authentication.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor uses local database authentication. Users are stored in the local database.
* A linux host with Docker CLI installed (Docker client).
* At least tow non-admin users. 

# Test Steps:

**NOTE:** The below non-admin user M should NOT be the same as the non-admin user in Test 2-23.

1. Assign an non-admin user M with system admin role and act as an admin user. 
2. Repeat all steps in Test 2-23.

# Expected Outcome:

* A user with system admin role can perform all operations the same as the admin user. 
* Same as Test 2-23.

# Possible Problems:
None