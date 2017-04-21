Test 8-03 - Update self-registration
=======

# Purpose:

To verify that an admin user can update self-registration setting.

# References:

User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. Login UI as admin user.  
2. In configuration page, uncheck self registration.  
3. Save configuration and logout.  

# Expected Outcome:

* In login page, sign up link will disappear.

# Possible Problems:
None
