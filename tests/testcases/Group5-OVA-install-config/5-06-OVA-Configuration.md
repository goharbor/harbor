Test 5-06 - OVA Configuration
=======

# Purpose:

To verify that the settings of an OVA version of Harbor can be configured and re-configured after power-on/off.

# References:
User guide, Installation guide of Harbor OVA version.

# Environment:
* This test requires an OVA binary of Harbor.
* A vCenter, at least an ESX host, and a network that supports DHCP.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

NOTE:
This cases is used for verify ova arguments not been tested in other tests,Harbor a and Harbor b should be replaced by longer and meaningful names.
1. From vSphere Web Client, import Harbor's OVA onto an ESX host.
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.
3. Power on the imported OVA.
4. Open a console from vSphere client and login as root.
5. Run `ovfenv` to view arguments.
6. Run `docker exec -i -t harbor-db mysql -u root -p` to login database.
7. Input database password to login.
8. Login harbor ui as admin user.
9. Poweroff the virtual machine.
10. Edit virtual machine settings.
11. Modify arguments in vapp propertities.
12. Power on the virtual machine.
13. Repeat Step 4-8.
14. Import a ova,config protocol use https and set verify remote cert on named harbor a.
15. Import another ova use selfsigned cert named harbor b.
16. Login harbor a as admin.
17. Click admin options and select syatem management.
18. Click add a destination.
19. Input information of harbor b.
20. Click test connection.
21. Poweroff harbor a.
22. Edit a settings turn off verify remote cert.
23. Power on harbor a.
24. Repeat add destination and test connection.

# Expected Outcome:

In step5 should see arguments has been set.
In step7 user shold be able to login database as root user.
In step13 user can not login vm database and harbor ui with new passwords,but other settings change are effected.
In step20 will get a x509 error.
In step24 test will sucessful.

# Possible Problems:
None
