Test 9-14 User pull unsigned images(LDAP images)
=======

# Purpose:

To verify whether user can pull unsigned images with content trust enabled.

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a LDAP or AD server.
* A Linux host with Docker CLI(Docker client)installed.
* A non-admin user that has at least one project as project admin.

# Test Steps:

Same as Test 9-04 except that users are from LDAP/AD.

# Expected Outcome:

* Same as Test 9-04.

# Possible Problems:
None
