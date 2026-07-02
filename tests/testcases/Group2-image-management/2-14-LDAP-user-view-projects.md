Test 2-14 - User View Projects
=======

# Purpose:

To verify that a non-admin user can view projects.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is configured with LDAP/AD authentication. Users are stored in an external LDAP or AD directory.
* There is at least a non-admin user. 
* The user has at least 3 private projects.
* The registry has at least 3 public repositories.

# Test Steps:

Same as Test 2-04 except that users are from LDAP/AD.

# Expected Outcome:
* Same as Test 2-04.

# Possible Problems:
None