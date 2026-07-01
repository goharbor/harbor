Test 9-11 User push signed images
=======

# Purpose:

To verify user can sign and push images

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.  
* Harbor is configured with LDAP/AD authentication. Users are stored in an external LDAP or AD directory.  
* A Linux host with Docker CLI(Docker client) installed.

# Test Steps:

Same as Test 9-01 except that users are from LDAP/AD.

# Expected Outcome:

* Same as Test 9-01.

# Possible Problems:
None
