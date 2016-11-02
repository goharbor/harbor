Test 1-08 - Admin User Log In and Log Out (DB Mode or LDAP mode)
=======

# Purpose:

To verify that an admin user can log in and log out.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database or LDAP server. 
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. The admin user logs in to the UI by username.
2. The admin user logs out from the UI.
3. Use the incorrect password of the admin user to log in to the UI and check the error message.
4. On a Docker client host, use `docker login <harbor_host>` command to verify the admin user can log in.
5. Use `docker login <harbor_host>` command to log in with incorrect password.  


# Expected Outcome:
* The admin user can log in/out successfully in Step 1 & 2. 
* The error message in Step 3 should not show which input value is incorrect. It should only display the username and password combination is incorrect.
* Docker client can log in in Step 4.
* Docker client fails to log in in Step 5.

# Possible Problems:
None