Test 1-09 - Admin User Create, Delete and Recreate a User(DB Mode)
=======

# Purpose:

To verify that an admin user can create/delete/recreate a user when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. 
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. The admin user logs in to the UI.
2. The admin user creates a user from the UI.
3. On a different browser, log in as the newly created user.
4. The user views his/her own account settings.
5. The user create two projects in the UI.
6. On a Docker client host, use `docker login <harbor_host>` command to verify the user can log in.
7. The admin user deletes the user from the UI.
8. When clicking on any link on the page, the deleted user's session on the different browser should be redirected to the login page and logged out.  
9. On a Docker client host, use `docker login <harbor_host>` command to verify the user cannot log in.
10. The admin user re-creates a user with the same username of the deleted user.
11. On a different browser, log in as the re-created user.
12. The user views his/her own account settings.
13. The user views "My Projects" to see what projects he/she owns.
14. On a Docker client host, use `docker login <harbor_host>` command to verify the re-created user can log in.
15. The admin user view logs from the dashboard and should see two items of two project creation of the deleted user. The user has special number associated with him/her username.

# Expected Outcome:
* The newly created user can log in successfully in Step 3. 
* The newly created user can view his/her own settings as entered by the admin in Step 4. 
* The newly created user can create project successfully in Step 5. 
* The admin should be able to re-create the user in Step 10.
* The re-created user should be able to log in, however, the previous projects no longer belong to him/her (Step 11-13).
* Docker client logs in successfully in Step 14.
* Should see special user id in logs in Step 15.

# Possible Problems:
None