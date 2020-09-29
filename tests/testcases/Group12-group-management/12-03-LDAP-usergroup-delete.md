Test 12-03 LDAP Usergroup Delete
=======

# Purpose

To verify admin user can delete an LDAP group

# References:

User guide

# Environments:

* This test requires that a Harbor instance is running and available.
* An LDAP server is running and available, and enabled memberof overlay feature
* LDAP group config parameter are configured.
    1. ldap_group_basedn
    1. ldap_group_filter
    1. ldap_gid 
    1. ldap_group_scope

# Test Steps:

1. Login UI as admin user.
2. In `Administration->User Group` page, Add an LDAP with a valid group DN with group name.
3. In Project Member of library, assign this user group with a developer role to this user group.
4. In `Administration->User Group` page, Delete the user group with a different name.
5. Check Project Member of library, make sure there is no role for this user group.

# Expected Outcome:

* In step4 the user group is deleted, and all its project member information is removed too.

# Possible Problem:
None