Test 5-05 - OVA LDAP integration
=======

# Purpose:

To verify that an OVA version of Harbor can work with AD and LDAP. The LDAP/AD configuration can be updated after power-off/on.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:
* This test requires an OVA binary of Harbor.
* A vCenter, at least an ESX host, and a network that supports DHCP.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. From vsphere web client, import harbor's ova onto an esx host.
2. In the deployment wizard, enter different passwords of linux root user, harbor admin user and mysql root user.
3. Set auth mode to ldap_auth.
4. Enter necessary ldap settings.
5. Power on the imported ova.
6. A ldap user login the ui by username.
7. On a docker client host, use `docker login <harbor_host>` command to log in as ldap user.
8. Power off the virtual machine.
9. Right click on the VM and select "Edit Settings".
10. Modify ldap settings.
11. Power on the virtual machine.
12. A user of new ldap login the UI by username.
13. On a Docker client host, use `docker login <harbor_host>` command to log in as new ldap user.

# Expected Outcome:

* In Step6 and 7,ldap user can login sucessful.
* In step12 and 13,new ldap user can login sucessful.

# Possible Problems:
None
