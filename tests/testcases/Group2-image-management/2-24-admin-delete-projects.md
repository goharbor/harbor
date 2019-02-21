Test 2-24 - Admin User Delete Projects (DB Mode)
=======

# Purpose:

To verify that an admin user can delete other user's projects when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that two(2) Harbor instances are running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user. 

# Test Steps:

**NOTE:**  
* In below test, user A is non-admin user. User A and project X should be replaced by longer and meaningful names.
* Must use two kinds of browsers at the same time to ensure independent sessions. For example, use Chrome and Firefox, or Chrome and Safari. 
* DO NOT use the same browser to log in two users in different windows(tabs).

1. Log in to UI as user A (non-admin).
2. Create a project X so that the user has the project admin role.
3. On a Docker client, log in as User A and run `docker push` to push an image to project X, e.g. projectX/myimage:v1.
4. Run `docker pull` to verify images can be pulled successfully.
5. In UI, log out user A.
6. Log in as admin user.
7. Delete project X. (should fail with errors)
8. Delete all images of project X. 
9. Delete project X. 
10. As an admin user, view the log in dashboard and should see delete operation of project X and its images.

# Expected Outcome:
* Step 7, deleting project X should fail because there are images under it.
* Step 8-9, deleting images of project X and then deleting project X should succeed.
* Step 10, there should be logs for deletion and creation of project X.

# Possible Problems:
None