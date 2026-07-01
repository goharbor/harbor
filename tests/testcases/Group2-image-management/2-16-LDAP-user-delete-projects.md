Test 2-16 - User Delete Projects
=======

# Purpose:

To verify that a non-admin user can delete projects.

# References:
User guide

# Environment:
* This test requires that two(2) Harbor instances are running and available.
* Harbor is configured with LDAP/AD authentication. Users are stored in an external LDAP or AD directory.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user. 

# Test Steps:

Same as Test 2-06 except that users are from LDAP/AD.

# Expected Outcome:
* Same as Test 2-06.

# Possible Problems:
None