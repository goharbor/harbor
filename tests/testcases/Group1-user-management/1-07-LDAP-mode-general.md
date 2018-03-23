Test 1-07 - LDAP Mode general functions
=======

# Purpose:

To verify that Harbor's UI works properly in LDAP mode.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an AD or LDAP server. (auth_mode is set to **ldap_auth** .) 
* An Active Directory (AD) or LDAP server has been set up and it has a few users available for testing.

# Test Steps:

1. The login page should not have "sign up" button. There is no need to allow self-registration.
2. Log in as a non system admin user(LDAP/AD) to the UI, he/she should NOT see the option of changing password or updating account settings.
3. Log out the user. 
4. Log in as a system admin user to the UI, he/she should see the option of changing his/her own password or updating account settings.
5. The system admin user should NOT see the option of adding a new user.
6. The system admin user should see above LDAP/AD user in the list. 
7. From the list, the system admin assigns system admin role to an AD/LDAP user A.
8. On a different browser(e.g. if the admin logs in using Chrome, then choose Safari or FireFox ), log in as the AD/LDAP user A to verify that user A has admin privilege.
9. From the list, the system admin removes system admin role from user A.
10. On a different browser, refresh the UI to verify user A has no admin privilege any more.
11. The system admin user deletes an AD/LDAP user in Harbor. **NOTE:** The user can log in again to regain access, however, all the previous projects he/she is a member of are lost.
To really disable a user's login, the user must be removed or disabled in AD or LDAP.  

# Expected Outcome:
* As described in steps 1-6.
* A LDAP/AD user can be assigned or removed admin role in Step 7-10.
* The user can be deleted successfully in Step 11. 

# Possible Problems:
None