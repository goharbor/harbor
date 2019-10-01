Test 5-02 - OVA Reboot
=======

# Purpose:

To verify that an OVA version of Harbor can be rebooted. Its configuration remains unchanged and should work the same way as before a reboot.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:
* This test requires an OVA binary of Harbor.
* A vCenter, at least an ESX host, and a network that supports DHCP.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. From vSphere Web Client, import Harbor's OVA onto an ESX host.
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.
3. Leave the networking settings blank.
4. Configure email settings. (can be mail server that does not exist )
5. Power on the imported OVA.
6. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)
7. Open a browser and enter http://VM_IP_address.
8. Log in as admin user of Harbor.
9. Create a new project.
10. On a Docker client host, use `docker login <harbor_host>` command to log in as the admin user.
11. Run some `docker push` and `docker pull` commands to push images to the project and pull from the project.
12. On vSphere, open the console of Harbor's VM, log in as root user using the password entered during deployment.
13. On vCenter Web Client, reboot the VM. (soft reboot)
14. After the VM starts up, repeat Step 7-12, should work the same.
15. Power off the VM, and the power it on. (hard reboot)
16. After the VM starts up, repeat Step 7-12, should work the same.
17. On vSphere, open the console of Harbor's VM, log in as root user, type `ovfenv` command to verify environment variables are the same as those entered during deployment.


# Expected Outcome:
* In Step 1-12, everything should work without errors. The passwords entered during deployment should work in Step 7, 9 and 11.
* In Step 14, 16, the VM should work the same before the reboot or power-off.
* Step 17, environment variables should remain unchanged as those entered during deployment.

# Possible Problems:
None