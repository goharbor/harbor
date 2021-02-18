Test 2-22 - Admin User Search Projects (DB Mode)
=======

# Purpose:

To verify that a non-admin user can search projects when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user. 

# Test Steps:

**NOTE:**  
In below test, user A is non-admin user. User A should be replaced by a longer and meaningful name.

1. Log in to UI as user A (non-admin).
2. Create at least 5 projects with different names.
3. Log out and log in again as admin user.
4. Search projects with keywords to see if projects of user A can be matched by criteria.
5. Click on a few projects to toggle publicity on and off.
6. Check in "Public Projects" to verify the projects with publicity on can be listed.

# Expected Outcome:
* As described in step 4,6.

# Possible Problems:
None