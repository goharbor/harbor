Test 2-01 - User Create Project (DB Mode)
=======

# Purpose:

To verify that a non-admin user can create projects when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least two non-admin users are in Harbor. 

# Test Steps:

**NOTE:**  
In below test, user A and B are non-admin users. User A, B and project X, Y should be replaced by longer and meaningful names.

1. Log in to UI as user A (non-admin).
2. Create a new project X with publicity is off (default).
3. Create another new project Y with publicity is on .
4. While keeping user A logging on, in another browser log in as user B (non-admin).
5. User B checks his/her public projects to see if project Y is listed and project X is not listed.
6. User A changes project X's publicity to on, project Y's publicity to off.
7. User B refreshes his/her public projects to see if project X is listed and project Y is not listed.
8. On a Docker client host, use `docker login <harbor_host>` to log in as user A. 
9. User A runs `docker push` to push an image to project X, push an image to project Y.
10. User A checks in the browser that the images had been successfully pushed to project X and Y.
11. On a Docker client host, use `docker login <harbor_host>` to log in as user B. 
12. User B runs `docker pull` to an image of project X, and an image of project Y. 

# Expected Outcome:
* Step 5, user B should see project Y is listed and project X is not listed. 
* Step 7, user B should see project X is listed and project Y is not listed. 
* Step 9,10, images should be pushed to project X and Y successfully and can be viewed from UI.
* Step 11,12, user B can pull the image of project X, cannot pull from project Y.

# Possible Problems:
None