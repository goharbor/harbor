Test 12-01 LDAP Usergroup Add
=======

# Purpose

To verify admin user can add an LDAP group

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
1. In `Administration->User Group` page, Add an LDAP with a valid group DN with group name.
   ### Expected Result
      * The user group should be created with specified name.
1. In `Administration->User Group` page, Add an LDAP with a non-exist group DN 
   ### Expected Result
      * The user group can not be created
1. In `Administration->User Group` page, Add an LDAP with a group DN which already exist, but with different name.
   ### Expected Result
      * The user group is renamed to new user group name.
1. In `Administration->User Group` page, Add an LDAP with a valid group DN without group name.
   ### Expected Result
      * The user group is created and named with the same name in LDAP.
1. Change the configure parameter ldap_group_basedn to another DN, so that the LDAP user group is outside the base DN. 
1. In `Administration->User Group` page, Add an LDAP with a valid group DN but outside the base DN.
   ### Expected Result
      * The user group can not be created
1. Change ldap_group_scope from 2 to 0, so that the LDAP group can not be found with the current scope.
1. In `Administration->User Group` page, Add an LDAP with a valid group DN but can not be searched.
   ### Expected Result
      * The user group can not be created
1. Change ldap_group_filter to with a specified filter, so that it can filter out the current group DN.
1. In `Administration->User Group` page, Add an LDAP with a valid group DN but this group DN is filtered
   ### Expected Result
      * the user group can not be created.
1. Change ldap_gid with another attribute other than cn
1. In `Administration->User Group` page, Add an LDAP with a valid group DN, check the user group name.
   ### Expected Result
      * The user group is created, the group name is named by specified attributed.


