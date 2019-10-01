Test 8-02 -Update-project-restrict
=======

# Purpose:

To verify that an admin user can update project restrict.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. Login UI as admin user.
2. In configuration page, change Project Creation Restriction from everyone to admin only.
3. Logout admin user.
4. Login as non-admin user.
5. Try to add a project.

# Expected Outcome:

* In step5, non-admin user will not see the add project button.

# Possible Problems:
None
