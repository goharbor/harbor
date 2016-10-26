Test 3-12 - Admin Search Project Members (DB Mode)
=======

# Purpose:

To verify that an admin user can search members of a project when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least five(5) non-admin users are in Harbor. 
* At least 5 members in a project that admin user is not a member of.

# Test Steps:

**NOTE:**  
* In below test, user A, B are non system admin users. User A, B and project X should be replaced by longer and meaningful names.
* MUST use two kinds of browsers at the same time to ensure independent sessions. For example, use Chrome and Firefox, or Chrome and Safari. 
* DO NOT use the same browser to log in two users at the same time in different windows(tabs).

1. Log in to UI as admin.
2. Look for an existing project X that admin is not a member of, project X should have at least 5 members. 
3. Search members of project X using different criteria (keywords). 


# Expected Outcome:
* Step 3, admin should see results based on search keywords. 

# Possible Problems:
None