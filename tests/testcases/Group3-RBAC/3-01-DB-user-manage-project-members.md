Test 3-01 - Manage Project Members (DB Mode)
=======

# Purpose:

To verify that a non system admin user can add members of various roles to a project when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least three non-admin users are in Harbor.

# Test Steps:

**NOTE:**
* In below test, user A, B and C are non system admin users. User A, B, C and project X should be replaced by longer and meaningful names.
* MUST use two kinds of browsers at the same time to ensure independent sessions. For example, use Chrome and Firefox, or Chrome and Safari.
* DO NOT use the same browser to log in two users at the same time in different windows(tabs).

1. Log in to UI as user A (non-admin).
2. Create a new project X with publicity is off (default).
3. Verify that user A cannot change his/her own role of project X.
4. On a Docker client host, use `docker login <harbor_host>` to log in as user A.
5. Push an image to project X.
6. On the Docker client host, log out user A and log in as user B.
7. Use `docker pull` to pull the image from project X. (should fail)
8. Use `docker push` to push an image to project X. (should fail)
9. Keeps user A's UI session logging on, in a different browser log in as user B.
10. User B views "My Projects", should not see project X.

11. In user A's UI, add user B to project X as a guest.
12. In user B's UI, check project X is in "My projects" and view members and images of project X.
13. Verify that user B cannot change role of other members in project X.
14. Verify that user B cannot add a new member to project X.
15. On a Docker client host, user B again pulls the image from project X.
16. user B pushes a new image to project X. (should fail)
17. In user A's UI, update user B to developer role of project X.
18. In user B's UI, check project X is in "My projects" and view members and images of project X.
19. Verify that user B cannot change role of other members in project X.
20. Verify that user B cannot add a new member to project X.

21. On a Docker client host, user B pushes a new image to project X.
22. In user A's UI, update user B to project admin role of project X.
23. In user B's UI, check project X is in "My projects" and view members and images of project X.
24. Verify that user B can add a new member user C to project X.
25. Verify that user B can change role of user C in project X.
26. On a Docker client host, user B pushes a new image to project X.
27. In user A's UI, remove user B from project X.
28. On a Docker client host, user B pulls an image of project X. (should fail)
29. On a Docker client host, log out user B and log in as user C.
30. User C pulls an image of project X.

# Expected Outcome:

* Step 3, user B cannot change his/her own role, and cannot remove himself/herself from project X.
* Step 7, user B should fail to pull image from project X.
* Step 8, user B should fail to push image to project X.
* Step 9, make sure to keep A's session while using another browser to log in as user B.
* Step 10, user B should not see project X.
* Step 12-14, as described.
* Step 15, user B can pull the image of project X,
* Step 16, user B cannot push images to project X.
* Step 18-20, as described.
* Step 21, user B pushes an image successfully.
* Step 23-25, as described.
* Step 28, user B's pulling should fail
* Step 30, user C's pulling should succeed.

# Possible Problems:
None