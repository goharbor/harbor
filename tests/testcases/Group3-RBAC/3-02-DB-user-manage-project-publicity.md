Test 3-02 - Manage Project Publicity (DB Mode)
=======

# Purpose:

To verify that a non system admin user can change project publicity of a project when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least three non-admin users are in Harbor. 

# Test Steps:

**NOTE:**  
* In below test, user A, B are non system admin users. User A, B and project X should be replaced by longer and meaningful names.
* MUST use two kinds of browsers at the same time to ensure independent sessions. For example, use Chrome and Firefox, or Chrome and Safari. 
* DO NOT use the same browser to log in two users at the same time in different windows(tabs).

1. Log in to UI as user A (non-admin).
2. Create a new project X with publicity is on.
3. On a Docker client host, use `docker login <harbor_host>` to log in as user A. 
4. Push an image to project X.
5. On the Docker client host, log out user A and log in as user B. 
6. User B pulls an image from project X.
7. Keeps user A's UI session logging on, in a different browser log in as user B.
8. User B views "My Projects", should not see project X.
9. User B views "Public Projects", should see project X.
10. In user A's UI, change publicity of project X to off.
11. In user B's UI, user B checks "My Projects" and "Public Projects", should not see project X in both places.
12. On the Docker client host, user B pulls an image from project X. (should fail)
13. In user A's UI, change publicity of project X to on again.
14. User B views "My Projects", should not see project X.
15. User B views "Public Projects", should see project X.


# Expected Outcome:
* Step 6, user B's pulling image from project X should succeed. 
* Step 7, make sure to keep A's session while using another browser to log in user B.
* Step 8-11, as described.
* Step 12, user B's pulling should fail.
* Step 13-15, as described.

# Possible Problems:
None