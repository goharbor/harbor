Test 1-02 - User Log In and Log Out (DB Mode)
=======

# Purpose:

To verify that a non-admin user can log in and log out when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
**NOTE:** Use a non-admin user for this test case. Admin user has other test cases.

1. A non-admin user logs in to the UI by username.
2. The user logs out from the UI.
3. A non-admin user logs in to the UI by email.
4. The user logs out from the UI.
5. Use the incorrect password and username/email of the user to log in to the UI and check the error message.
6. On a Docker client host, use `docker login <harbor_host>` command to verify the user can log in by either the **username** or **email** . (check both)
7. Use `docker login <harbor_host>` command to log in with incorrect password by either the **username** or **email** .


# Expected Outcome:
* The user can log in via UI in Step 1 & 3, verify the dashboard and navigation bar are for a non-admin user. (should not see admin options)
* After the user logged out in Step 2 & 4, the login page will be displayed again.
* The error message in Step 5 should not show which input value is incorrect. It should only display the username(email) and password combination is incorrect.
* Docker client can log in in Step 6.
* Docker client fails to log in in Step 7.

# Possible Problems:
None