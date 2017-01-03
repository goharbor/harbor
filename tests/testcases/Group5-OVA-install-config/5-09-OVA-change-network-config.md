Test 5-09 - OVA Networking Settings After Deployment
=======

# Purpose:

To verify that an OVA version of Harbor can be configured to obtain IP address from DHCP or static IP after deployment.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:
* This test requires an OVA binary of Harbor.
* A vCenter, at least an ESX host, and a network that supports DHCP.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. From vSphere Web Client, import Harbor's OVA onto an ESX host.
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.
3. Leave network settings blank
4. Power on the imported OVA.
5. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)
6. Open a browser and enter https://VM_IP_address:5480.
7. Login as root user.
8. Change network address settings to static ip address.
9. Input static ip settings.
10. Save settings and reboot.
11. Wait a few minutes and open a browser and enter https://VM_IP_address:5480.
12. Login as root user and check the network status.
13. Import a new ova.
14. In the deployment wizard,set network address to static ip.
15. Power on the VM.
18. Open a browser and enter https://VM_IP_address:5480.
17. Login as root user.
18. Change address settings to dhcp.
19. Save settings and reboot.
20. Wait a few minutes and open a browser and enter https://VM_IP_address:5480.
21. Login as root user and check the network status.


# Expected Outcome:
* In Step 12,network status should changed from dhcp to static ip. 
* In Step 21,network status should changed from static ip to dhcp. 

# Possible Problems:
None
