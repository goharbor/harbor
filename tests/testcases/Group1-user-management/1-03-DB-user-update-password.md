Test 1-03 - User update password (DB Mode)
=======

# Purpose:

To verify that a non-admin user can update password when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
**NOTE:** Use a **non-admin** user for this test case. Admin user has other test cases.

1. A non-admin user logs in to the UI by **username**.
2. The user changes his/her own password.
3. The user logs out.
4. The same user logs in to the UI by **email** using the new password.
5. The user can log in using `docker login` command using the new password.
6. The user goes to the page to change his/her own password. Provide invalid values of input to see if validation works:  

* old password is incorrect
* password input does not compliant to password rule
* two passwords do not match

# Expected Outcome:
* Password can be changed in Step 2.
* User can log in using new password in Step 4.
* In Step 6, the user cannot change password due to various errors. Proper error message should be displayed.

# Possible Problems:
None