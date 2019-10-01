Test 8-01 -Update-authentication-mode
=======

# Purpose:

To verify that an admin user can update authentication mode.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
**NOTE:**
Before this test, make sure there is only admin user in the system.

1. Login UI as admin user.
2. In configuration page,change authentication mode from DB to LDAP or from LDAP to DB.
3. Save the configuration.
4. In ldap_auth mode, fill in the ldap info and click test ldap server.
5. Add or sign up a user.(For LDAP, login a user)
6. Change the configuration again.

# Expected Outcome:

* In step2, user can change authentication mode.
* In step4, if settings are correct, test will successful.
* In step6, user cannot change authentication mode.

# Possible Problems:
None
