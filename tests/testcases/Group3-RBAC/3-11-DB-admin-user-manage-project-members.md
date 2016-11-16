Test 3-11 - Admin Manages Project Members (DB Mode)
=======

# Purpose:

To verify that an admin user can add members of various roles to a project. Users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least three non-admin users are in Harbor. 
* At least one project that admin user is not a member of.

# Test Steps:

**NOTE:**  
* In below test, user A, B and C are non system admin users. User A, B, C and project X should be replaced by longer and meaningful names.
* MUST use two kinds of browsers at the same time to ensure independent sessions. For example, use Chrome and Firefox, or Chrome and Safari. 
* DO NOT use the same browser to log in two users at the same time in different windows(tabs).

1. Log in to UI as admin user.
2. Look for an existing project X that admin is not a member of. 
3. Add user A to project X as project admin.
4. Add user B to project X as developer.
5. Add user C to project X as guest.
6. On a Docker client host, use `docker login <harbor_host>` to log in as admin. 
7. Use `docker push` to push an image to project X. 
8. Use `docker pull` to pull the image from project X. 
9. On a Docker client host, log out admin and log in as user A. 
10. Use `docker push` to push an image to project X. 
11. Use `docker pull` to pull the image from project X.
 
12. On a Docker client host, log out user A and log in as user B. 
13. Use `docker push` to push an image to project X. 
14. Use `docker pull` to pull the image from project X. 
15. On a Docker client host, log out user B and log in as user C. 
16. Use `docker pull` to pull the image from project X. 
17. Use `docker push` to push an image to project X. (should fail)

18. Keeps admin's UI session logging on, in a different browser log in as user C.
19. In user C's UI, verify his/her role is guest of project X. 
20. In admin's UI, change user C's role to developer of project X.
21. In user C's UI, verify his/her role is developer of project X. 
22. On a Docker client host, log in as user C. 
23. Use `docker pull` to pull the image from project X. 
24. Use `docker push` to push an image to project X. 
25. In admin's UI, change user C's role to project admin of project X.
26. In user C's UI, verify his/her role is project admin of project X. 

27. In admin's UI, remove user C from project X.
28. Set project X's publicity to on.
29. On a Docker client host, log in as user C. 
30. Use `docker pull` to pull the image from project X. 
31. Use `docker push` to push an image to project X. (should fail)
32. Set project X's publicity to off.
33. On a Docker client host, log in as user C. 
34. Use `docker pull` to pull the image from project X. (should fail) 
35. Use `docker push` to push an image to project X. (should fail)

# Expected Outcome:

* Step 7,8,10,11,16 should succeed.
* Step 17 should report errors.
* Step 19,21, as described.
* Step 23,24 should succeed.
* Step 26 as described.
* Step 30 should succeed.
* Step 31 should fail.
* Step 34-35 should fail.

# Possible Problems:
None