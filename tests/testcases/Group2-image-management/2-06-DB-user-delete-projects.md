Test 2-06 - User Delete Projects (DB Mode)
=======

# Purpose:

To verify that a non-admin user can delete projects when users are managed locally by Harbor (DB mode).

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
* DO NOT use the same browser to log in two users at the same time in different windows(tabs).

1. Log in to UI as user A (non-admin).
2. Create a project X so that the user has the project admin role.
3. On a Docker client, log in as User A and run `docker push` to push an image to project X, e.g. projectX/myimage:v1.
4. Push an image with different name to project X, e.g. projectX/newimage:v1 .
5. Run `docker pull` to verify images can be pulled successfully.
6. In UI, delete project X directly. (should fail with errors)
7. While keeping the current user A logged on, in a different browser, log in as admin user. 
8. Under "Admin Options", create a replication policy of project X to another Harbor instance. (Do not need to activate this policy.)
9. Switch to the UI of User A, delete all images under project X.
10. In user A's UI, delete project X directly. (should fail with errors)
11. Switch to the UI of admin user, delete the replication policy of project X.
12. In user A's UI, delete project X. 
13. In user A's UI, recreate project X, 
14. On a Docker client, log in as User A and run `docker push` to push an image to project X, e.g. projectX/anotherimage:v1. The image name should not be the same as those deleted in previous steps.
15. Switch to the UI of admin user, view images under the re-created project X.
16. As an admin user, view the log in dashboard and should see delete and create operations of project X.

# Expected Outcome:
* Step 6, deleting project X should fail because there are images under it.
* Step 10, deleting project X should fail because there is an image replication policy under it.
* Step 12, deleting project X should succeed.
* Step 13, re-creation of project X should succeed.
* Step 14, push should succeed.
* Step 15, project X should contain newly pushed image. The old images should not be displayed.
* Step 16, there should be logs for delete and create of project X, notice the project name of the deleted operation is displayed differently.

# Possible Problems:
None