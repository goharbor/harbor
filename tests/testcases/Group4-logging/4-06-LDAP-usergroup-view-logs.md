Test 4-06 - User Views Logs (LDAP Mode)
=======

# Purpose:

To verify that a LDAP user group can views logs when users are managed externally by LDAP or AD (LDAP mode).

# References:
User guide

# Environment:

* This test requires that a Harbor instance is running and available.
* An LDAP server is running and available, and enabled memberof overlay feature.
* Harbor is set to authenticate against an LDAP or AD server. ( auth_mode is set to **ldap_auth** .) The user data is stored in an LDAP or AD server.
* A linux host with Docker CLI installed (Docker client).
* LDAP group config parameter are configured.
    1. ldap_group_basedn
    1. ldap_group_filter
    1. ldap_gid 
    1. ldap_group_scope   

# Test Steps:

1. Add group harbor_admin and create a user admin_user, admin_user is a member of harbor_admin
2. Login to UI with admin user, create a private project ldap_group_proj
3. Add a project member with ldap the LDAP DN of harbor_admin, with administrator role
4. Log in to the UI as the admin_user in docker client.
5. push/pull images to ldap_group_proj.
6. View the logs of the project. 
7. Try below search criteria to see if the search result is correct:

* push only
* pull only
* pull and push
* delete only
* all
* push and delete
* different date ranges 
* date range and push

# Expected Outcome:

* All operations in Step 5 should be logged.
* Logs can be viewed in Step 6, check if the time and operations are correct.
* Logs can be filtered in Step 6.

# Possible Problems:
None