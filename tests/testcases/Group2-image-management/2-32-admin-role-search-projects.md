Test 2-32 - Admin Role User Search Projects
=======

# Purpose:

To verify that a user with system admin role can search projects using local database authentication.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor uses local database authentication. Users are stored in the local database.
* A linux host with Docker CLI installed (Docker client).
* At least two non-admin users are in Harbor. 

# Test Steps:


**NOTE:** The below non-admin user M should NOT be the same as the non-admin user in Test 2-22.

1. Assign an non-admin user M with system admin role and act as an admin user. 
2. Repeat all steps in Test 2-22.

# Expected Outcome:

* A user with system admin role can perform all operations the same as the admin user. 
* Same as Test 2-22.

# Possible Problems:
None