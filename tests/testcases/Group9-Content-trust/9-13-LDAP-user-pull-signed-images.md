Test 9-13 User pull signed images(LDAP mode)
=======

# Purpose:

To verify whether user can pull signed images with content trust enabled.

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against an LDAP or AD server.
* A Linux host with Docker CLI(Docker client)installed.
* A non-admin user that has at least one proejct as project admin.

# Test Steps:

Same as Test 9-03 except that users are from LDAP/AD.

# Expected Outcome:

* Same as Test 9-03.

# Possible Problems:
None
