Test 4-01 - User Views Logs (DB Mode)
=======

# Purpose:

To verify that a non-admin user can views logs when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
1. On a Docker client host, use `docker login <harbor_host>` command to log in as a non-admin user. 
2. Run some `docker push` and `docker pull` commands to push images to the registry and pull from the registry.
3. Log in to the UI as the non-admin user.
4. Delete a few images from the project. 
5. View the logs of the project. 
6. Try below search criteria to see if the search result is correct:

* push only
* pull only
* pull and push
* delete only
* all
* push and delete
* different date ranges 
* date range and push

# Expected Outcome:
* All operations in Step 2 & 4 should be logged.
* Logs can be viewed in Step 5, check if the time and operations are correct.
* Logs can be filtered in Step 6.

# Possible Problems:
None