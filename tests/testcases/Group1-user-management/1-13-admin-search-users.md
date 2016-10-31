Test 1-13 - Admin User Search Users.
=======

# Purpose:

To verify that Admin user can search users.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database, LDAP or AD.

# Test Steps:

1. Log in as the admin user to the UI.
2. The admin user should see a list of users in "Admin Options". 
3. Enter different keywords or partial words to see if the user list can be filtered according to the criteria.

# Expected Outcome:
* As described in steps 1-3.

# Possible Problems:
None