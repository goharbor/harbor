Test 2-03 - User Create Multiple Projects
=======

# Purpose:

To verify that a non-admin user can create multiple projects using local database authentication.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor uses local database authentication. Users are stored in the local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user. 

# Test Steps:

**NOTE:**  
In below test, user A is non-admin user. User A should be replaced by a longer and meaningful name.

1. Log in to UI as user A (non-admin).
2. Create 16 or more projects so that the pagination control has multiple pages.
3. Go through multiple pages of the list and click on a few projects to see if pagination work properly.
4. Search projects with keywords to see if the list and pagination update accordingly.

# Expected Outcome:
* As described in step 3-4.

# Possible Problems:
None