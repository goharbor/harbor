Test 1-06 - AD (Active Directory) User Log In and Log Out (LDAP Mode)
=======

# Purpose:

To verify that a non-admin user can log in and log out when users are managed externally by an AD server (LDAP mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an AD server. (auth_mode is set to **ldap_auth** .) The user data is stored in an AD server.
* A linux host with Docker CLI installed (Docker client).
* An Active Directory (AD) server has been set up and it has a few users available for testing.

# Test Steps:

1. A user has **NEVER** logged in to Harbor. He/she logs in to the UI for the first time by his/her id in AD. The id could be the cn attribute (or what is configured in ldap_uid) of his/her AD user DN.
2. The user logs out from the UI.
3. The user logs in again to the UI, should not see his/her own account settings and cannot change password(need to go to LDAP/AD for this).
4. The user logs out from the UI.
5. Use the incorrect password and username of the user to log in to the UI and check the error message.
6. On a Docker client host, use `docker login <harbor_host>` command to verify the user can log in by username/password. 
7. Run `docker login <harbor_host>` command to log in with incorrect password of the user.  
8. Log in as a system admin to UI, go to "admin Options" and should see above AD user in the list. System admin can assign or remove system admin role of the above AD user.
9. Disable or remove the user in AD.
10. The user should no longer log in to UI or by Docker client.

# Expected Outcome:
* The user can log in to UI by AD id in Step 1 & 3, verify the dashboard and navigation bar are for a non-admin user. (should not see admin options)
* In Step 3, also verify that the user cannot update his/her account settings and cannot change password.
* After the user logged out in Step 2 & 4, the login page will be displayed again.
* The error message in Step 5 should not show which input value is incorrect. It should only display the username(email) and password combination is incorrect.
* Docker client can log in in Step 6.
* Docker client fails to log in in Step 7.
* AD user should have no difference from a user in local database in Step 8.
* The user's login should fail in Step 10.

# Possible Problems:
None