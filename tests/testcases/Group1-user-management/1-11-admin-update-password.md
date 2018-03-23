Test 1-11 - Admin User update password (DB Mode)
=======

# Purpose:

To verify that the admin user can update password.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. The admin user logs in to the UI by **username**.
2. The admin user changes his/her own password.
3. The admin user logs out.
4. The admin user logs in to the UI by **email** using the new password.
5. The admin user can log in using `docker login` command using the new password.

# Expected Outcome:
* Password can be changed in Step 2.
* User can log in using new password in Step 4.

# Possible Problems:
None