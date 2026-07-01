Test 3-12 - Admin Search Project Members
=======

# Purpose:

To verify that an admin user can search members of a project using local database authentication.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor uses local database authentication. Users are stored in the local database.
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