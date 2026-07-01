Test 2-13 - User Create Multiple Projects
=======

# Purpose:

To verify that a non-admin user can create multiple projects.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is configured with LDAP/AD authentication. Users are stored in an external LDAP or AD directory.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user. 

# Test Steps:

Same as Test 2-03 except that users are from LDAP/AD.

# Expected Outcome:
* Same as Test 2-03.

# Possible Problems:
None