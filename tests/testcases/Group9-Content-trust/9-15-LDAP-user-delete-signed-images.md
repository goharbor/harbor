Test 9-15 User delete signed images(LDAP mode)
=======

# Purpose:

To verify user can delete signed images with content trust enabled.

# References:
User guide

# Environment:

* This test requires a Harbor instance is running and available.  
* Harbor is set authenticate against an LDAP or AD server.  
* A Linux host with Docker CLI(Docker client)installed.  
* A non-admin user that has at least one project as project admin.

# Test Steps:

Same as Test 9-05 except that users are from LDAP/AD.

# Expected Outcome:

* Same as Test 9-05.

# Possible Problems:
None
