Test 2-11 - User Create Project
=======

# Purpose:

To verify that a non-admin user can create projects.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is configured with LDAP/AD authentication. Users are stored in an external LDAP or AD directory.
* A linux host with Docker CLI installed (Docker client).
* At least two non-admin users are in Harbor. 

# Test Steps:

Same as Test 2-01 except that users are from LDAP/AD.

# Expected Outcome:
* Same as Test 2-01.

# Possible Problems:
None