Test 1-10 - Admin User update account settings (DB Mode)
=======

# Purpose:

To verify that the admin user can update his/her account settings.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. The admin user logs in to UI.
2. The user changes his/her account settings, including email, full name and comments.
3. The user logs out.
4. The admin user logs in again using **new email**, and verify the account settings had been changed.

# Expected Outcome:
* Account settings can be changed in Step 2.
* User can log in using new email in Step 4 and the settings are the same as input in Step 2.

# Possible Problems:
None
