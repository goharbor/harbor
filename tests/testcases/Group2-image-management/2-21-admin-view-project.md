Test 2-21 - Admin View Projects (DB Mode)
=======

# Purpose:

To verify that an admin user can view all projects when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user is in Harbor. 

# Test Steps:

**NOTE:**  
In below test, user A is non-admin user. User A and project X, Y should be replaced by longer and meaningful names.

1. Log in to UI as user A (non-admin).
2. Create a new project X with publicity is off (default).
3. Create another new project Y with publicity is on.
4. On a Docker client host, use `docker login <harbor_host>` to log in as user A. 
5. Push an image to project X, push another image to project Y.
6. In UI, logs out and log in as admin user.
7. Check "Public Projects" to verify that project Y is listed and project X is not listed.
8. Verify that project "library" is listed.
9. Click project Y and view the newly pushed image in project Y.
10. Click "Users" to view the members of project Y. 
11. Check "My projects" to verify that both project X and project Y are listed.
12. Verify that project "library" is listed.
13. Click project X and view the newly pushed image in project X.
14. Click "Users" to view the members of project X. 

# Expected Outcome:
* Step 7, admin should see project Y is listed and project X is not listed. 
* Step 8, project library should be listed. 
* Step 9, the image pushed by user A should be listed under project Y.
* Step 10, user A should be project admin of project Y.
* Step 11, admin should see project X and Y are both listed. 
* Step 12, project library should be listed. 
* Step 13, the image pushed by user A should be listed under project X.
* Step 14, user A should be project admin of project X.

# Possible Problems:
None