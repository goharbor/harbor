Test 2-23 - Admin User Delete Images (DB Mode)
=======

# Purpose:

To verify that an admin user can delete images owned by other users when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user.

# Test Steps:

**NOTE:**
In below test, user A is non-admin user. User A and project X should be replaced by longer and meaningful names.

1. Log in to UI as user A (non-admin).
2. Create a project X so that the user has the project admin role.
3. On a Docker client, log in as User A and run `docker push` to push an image to the project X, e.g. projectX/myimage:v1.
4. Push an image with different name to project X, e.g. projectX/newimage:v1 .
5. Run `docker pull` to verify images can be pulled successfully.
6. In UI, log out user A's session.
7. Log in as admin user.
8. Under project X, delete the two images one by one.
9. On a Docker client, log in as User A and run `docker pull` to pull the two deleted images of project X.

# Expected Outcome:
* Step 8, admin user can delete images of project X.
* Step 9, docker client should report errors.

# Possible Problems:
None