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
This cases are used for verify OVA arguments not been tested in other tests, Harbor A and Harbor B should be replaced by longer and meaningful names.  
1. From vSphere Web Client, import Harbor's OVA onto an ESX host.  
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.  
3. Power on the imported OVA.  
4. Open a console from vSphere client and login as root.  
5. Run `ovfenv` to view arguments.  
6. Run `docker exec -i -t harbor-db mysql -u root -p` to login database.  
7. Input database password to login.  
8. Login harbor UI as admin user.  
9. Power off the virtual machine.  
10. Edit virtual machine settings.  
11. Modify arguments in vapp properties.  
12. Power on the virtual machine.  
13. Repeat Step 4-8.  
14. Import a OVA, config protocol use https and set verify remote cert on named Harbor A.  
15. Import another OVA use self-signed cert named Harbor B.  
16. Login Harbor A as admin.  
17. Click admin options and select system management.  
18. Click add a destination.  
19. Input information of Harbor B.  
20. Click test connection.  
21. Power off Harbor A.  
22. Edit a settings turn off verify remote cert.  
23. Power on Harbor A.  
24. Repeat add destination and test connection.

# Expected Outcome:

* In step5 should see arguments has been set.
* In step7 user should be able to login database as root user.
* In step13 user cannot login VM database and harbor UI with new passwords, but other settings change are effected.
* In step20 will get a x509 error.
* In step24 test will successful.

# Possible Problems:
None
