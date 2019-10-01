Test 12-02 LDAP Usergroup Update
=======

# Purpose

To verify admin user can update an LDAP group

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
3. In `Administration->User Group` page, Update the user group with a different name.


# Expected Outcome:


* In step3 the user group name is updated

# Possible Problem:
None