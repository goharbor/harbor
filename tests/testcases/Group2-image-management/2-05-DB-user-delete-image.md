Test 2-05 - User Delete Images (DB Mode)
=======

# Purpose:

To verify that a non-admin user can delete images when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user.

# Test Steps:

**NOTE:**
In below test, user A is non-admin user. User A and project X, Y should be replaced by longer and meaningful names.

1. Log in to UI as user A (non-admin).
2. Create a project X so that the user has the project admin role.
3. On a Docker client, log in as User A and run `docker push` to push an image to the project X, e.g. projectX/myimage:v1.
4. Push a second image with different tag to project X, e.g. projectX/myimage:v2 .
5. Push an image with different name to project X, e.g. projectX/newimage:v1 .
6. Run `docker pull` to verify images can be pulled successfully.
7. In UI, delete the three images one by one.
8. On a Docker client, log in as User A and run `docker pull` to pull the three deleted images of project X.
9. In UI, delete project X.
10. Run `docker pull` to pull the three deleted images of the project X.

# Expected Outcome:
* Step 7, images should be deleted successfully.
* Step 9, project X should be deleted successfully.
* Step 8,10, docker client should report error message.

# Possible Problems:
None