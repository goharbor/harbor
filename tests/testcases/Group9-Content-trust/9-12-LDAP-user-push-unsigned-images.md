Test 9-12 User push unsigned images(LDAP mode)
=======

# Purpose:

To verify UI will difference unsigned images from signed.

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.  
* Harbor is set to authenticate against a LDAP or AD server.   
* A Linux host with Docker CLI(Docker client)installed.  
* A non-admin user that has at least a project as project admin.

# Test Steps:

Same as Test 9-02 except that users are from LDAP/AD.

# Expected Outcome:

* Same as Test 9-02.

# Possible Problems:
None
