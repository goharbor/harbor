Test 2-04 - User View Projects
=======

# Purpose:

To verify that a non-admin user can view projects using local database authentication.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor uses local database authentication. Users are stored in the local database.
* There is at least a non-admin user. 
* The user has at least 3 private projects.
* The registry has at least 3 public repositories.

# Test Steps:

**NOTE:**  
In below test, user A is non-admin user. User A should be replaced by a longer and meaningful name.

1. Log in to UI as user A (non-admin).
2. Create at least 3 projects if he/she has less than 3 projects.
3. Switch a few times between "My Projects" and "Public projects" tab, view listed projects and click to check details of images.
4. Check logs of projects in "My Projects".

# Expected Outcome:
* Step 3, verify the information listed of projects are correctly displayed, such as creation time and role.
* Step 4, should see logs of the project.

# Possible Problems:
None