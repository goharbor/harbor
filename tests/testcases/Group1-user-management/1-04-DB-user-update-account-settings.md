Test 1-04 - User update account settings (DB Mode)
=======

# Purpose:

To verify that a non-admin user can update his/her account settings when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
**NOTE:** Use a **non-admin** user for this test case. Admin user has other test cases.

1. A non-admin user logs in to UI.
2. The user changes his/her account settings, including email, full name and comments.
3. The user logs out.
4. The same user logs in again using **new email**, and verify the user's account settings had been changed.
5. The user goes to the page to change his/her settings again. Provide invalid values of input to see if validation works:  

* email formatting
* very long email address string
* required fields are empty

# Expected Outcome:
* Account settings can be changed in Step 2.
* User can log in using new email in Step 4 and the settings are the same as input in Step 2.
* In Step 5, the user cannot change account settings due to various errors. Proper error message should be displayed.

# Possible Problems:
None