Test 8-01 -Update-authentication-mode
=======

# Purpose:

To verify that an admin user can update authentication mode(This will work only the first time).  

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. Login UI as admin user.  
2. In configuration page,change authentication mode from DB to LDAP or from LDAP to DB.  
3. Save the configuration.  
4. Add or sign up a user.(For LDAP, login a user)
5. Change the configuration again.  

# Expected Outcome:

* In step2, user can change authentication mode.  
* In step5, user cannot change authentication mode.

# Possible Problems:
None
