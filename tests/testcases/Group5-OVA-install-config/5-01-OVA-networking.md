Test 5-01 - OVA Networking Settings During Deployment
=======

# Purpose:

To verify that an OVA version of Harbor can be configured to obtain IP address from DHCP or static IP.

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
4. Power on the imported OVA.
5. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)
6. Open a browser and enter http://VM_IP_address.
7. Log in as admin user of Harbor.
8. Create a new project.
9. On a Docker client host, use `docker login <harbor_host>` command to log in as the admin user.
10. Run some `docker push` and `docker pull` commands to push images to the project and pull from the project.
11. On vSphere, open the console of Harbor's VM, log in as root user using the password entered during deployment.
12. Check the network IP address of the VM.
13. From vSphere Web Client, import a new Harbor's OVA onto an ESX host.
14. Repeat Step 2-12, except in Step 3, enter static network settings of the OVA, such as static IP, DNS, gateway, hostname.


# Expected Outcome:
* In Step 1-11, everything should work without errors. The passwords entered during deployment should work in Step 7, 9 and 11.
* In Step 14, basically is to test if static networking works for the OVA. The outcome should be the same as Step 1-12, except the networking settings are from the deployment wizard.
Verify the IP address, hostname, DNS are correctly set inside the VM.

# Possible Problems:
None