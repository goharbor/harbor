Test 3-03 - Search Project Members (DB Mode)
=======

# Purpose:

To verify that a non system admin user can search members of a project when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least five(5) non-admin users are in Harbor. 

# Test Steps:

**NOTE:**  
* In below test, user A, B are non system admin users. User A, B and project X should be replaced by longer and meaningful names.
* MUST use two kinds of browsers at the same time to ensure independent sessions. For example, use Chrome and Firefox, or Chrome and Safari. 
* DO NOT use the same browser to log in two users at the same time in different windows(tabs).

1. Log in to UI as user A (non-admin).
2. Create a new project X.
3. Add user B as a project admin role of project X. 
4. Add 3 more members to project X with various roles, such as developer, guest. 
5. Keeps user A's UI session logging on, in a different browser log in as user B.
6. In user B's UI, search members of project X using different criteria (keywords). 
7. In user A's UI, change user B's role to developer of project X.
8. In user B's UI, search members of project X using different criteria. 
9. In user A's UI, change user B's role to guest of project X.
10. In user B's UI, search members of project X using different criteria. 


# Expected Outcome:
* Step 6,8,10, user B should see results based on search keywords. 

# Possible Problems:
None