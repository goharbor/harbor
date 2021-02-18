Test 1-01 - User Registration (DB Mode)
=======

# Purpose:

To verify that a non-admin user can register an account (signup) when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
1. On the Harbor's home page, click "Sign Up"
2. Enter user information to register a user.
3. Use the username of the newly registered user to log in to the UI.
4. Log out from the UI.
5. Use the email of the newly registered user to log in to the UI.
6. On a Docker client host, use `docker login <harbor_host>` command to verify the user can log in by either the **username** or **email** . (verify both) 
7. Log out from the UI and register another new user. Try to provide invalid values of input to see if validation works: 


* username is the same as an existing user
* username is very long in length
* wrong email formatting
* email is very long in length
* password input does not compliant to password rule
* two passwords do not match


# Expected Outcome:
* A new user created in step 2. 
* The new user can log in via UI in Step 3 and Step 5.
* The new user can log in via docker client in Step 6 by email and username.
* Invalid input during sign up can be rejected, proper error messages can be displayed in Step 7.

# Possible Problems:
None