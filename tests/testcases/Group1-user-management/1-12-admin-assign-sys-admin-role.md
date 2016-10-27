Test 1-12 - Admin User Assign and Remove Admin Role to a User.
=======

# Purpose:

To verify that Admin user can assign/remove admin role to/from a user.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.

# Test Steps:

1. Log in as the admin user to the UI.
2. The admin user should see a list of users in "Admin Options". 
3. From the list, the admin user assigns system admin role to a user A.
4. On a different browser(e.g. if the admin logs in using Chrome, then choose Safari or FireFox ), log in as user A to verify that user A has admin privilege.
5. From the list, the admin user removes system admin role from user A.
6. On a different browser, refresh the UI to verify user A has no admin privilege any more.

# Expected Outcome:
* As described in steps 1-6.

# Possible Problems:
None