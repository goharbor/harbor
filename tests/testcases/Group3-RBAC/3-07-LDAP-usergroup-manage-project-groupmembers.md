Test 3-07 - LDAP usergroup manage project group members
=======
# Purpose:

To verify LDAP group can be assigned a role in project member

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.
* An LDAP server is running and available, and enabled memberof overlay feature
* Harbor is set to authenticate against an LDAP or AD server. ( auth_mode is set to **ldap_auth** .) The user data is stored in an LDAP or AD server.
* LDAP group config parameter are configured.
    1. ldap_group_basedn
    1. ldap_group_filter
    1. ldap_gid
    1. ldap_group_scope
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. Create group  harbor_guest, harbor_dev, harbor_admin in LDAP.
1. Create LDAP user guest_user, dev_user, admin_user in LDAP.
    Assign add group following members
      * harbbor_guest --- guest_user, admin_user.
      * harbor_dev    --- dev_user.
      * harbor_admin  --- admin_user.

1. Login as admin user, create private project proj_group_test
1. Add following group with the roles to proj_group_test
    * harbor_guest  --- guest, add this member with LDAP Group DN directly: cn=harbor_guest,ou=groups,dc=example,dc=com.
    * harbor_dev    --- developer, create user group with LDAP group DN directly: cn=harbor_developer,ou=groups,dc=example,dc=com
    * Add a user group: group DN: cn=harbor_admin,ou=groups,dc=example,dc=com, with name harbor_admin,
    * Add project member, select existing user group harbor_admin, assign role administrator.
1. Login user guest_user, dev_user, admin_user in web console. all of them can see the proj_group_test.

   ### Expected Results:

   * All LDAP users guest_user, dev_user, admin_user can login and see the proj_group_test in web console.
   * guest_user has guest role in proj_group_test
   * dev_user has developer role in proj_group_test
   * admin_user has administrator role in proj_group_test

1. Login user guest_user, dev_user, admin_user in command line. try to push pull images.

   ### Expected Results:

   * All LDAP users can login to harbor in command line.
   *  guest_user -- can pull images
   *  dev_user   -- can pull/push images
   *  admin_user -- can pull/push images

1. Remove admin_user from LDAP group harbor_admin, login again with admin_user. check the role in project proj_group_test
   ### Expected Results:
   *  After remove harbor_admin membership, the admin_user should have guest role in project proj_group_test.
1. Remove admin_user from LDAP group harbor_guest, login again with admin_user, check the role in project
   ### Expected Results:
   *  After remove harbor_guest membership, the admin_user can not see the project proj_group_test.