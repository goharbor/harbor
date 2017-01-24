Test 5-04 - OVA Uses HTTPS
=======

# Purpose:

To verify that an OVA version of Harbor can set up for HTTPS.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:
* This test requires an OVA binary of Harbor.
* A vCenter, at least an ESX host, and a network that supports DHCP.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

NOTE:  
In below test, user A should be replaced by longer and meaningful names.  
1. From vSphere Web Client, import Harbor's OVA onto an ESX host.  
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.  
3. In the deployment wizard, set protocol to http.  
4. Power on the imported OVA.  
5. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)  
6. Open a browser and enter http://VM_IP_address.  
7. Log in as admin user of Harbor.  
8. Click username and select about.  
9. Check if there is a default ca certificate download link.  
10. Power off the virtual machine.  
11. Edit the virtual machine settings.  
12. Set protocol to https and leave cert and key blank.  
13. Power on the virtual machine.  
14. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)  
15. Open a browser and enter https://VM_IP_address.  
16. Log in as admin user of Harbor.  
17. Click username and select about.  
18. Click default ca certificate download link and save the file.  
29. login as user A.  
20. Click username and select about.  
21. Check if there is a default ca certificate download link.  
22. Login as admin user.  
23. Modify user A to administrators.  
24. Logout admin and login user A.  
25. Check if there is a default ca certificate download link.  
26. On a Docker client, Run `docker login <harbor_host> command` to login.  
27. Put the downloaded certificate file into /etc/docker/cert.d/VM_IP_address(or FQDN)/.  
28. Run `docker login <harbor_host> command` to login.  
29. Power off the virtual machine.  
30. Edit the virtual machine settings.  
31. Set protocol to https and input cert and key.  
32. Power on the virtual machine.  
33. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)  
34. Open a browser and enter https://VM_IP_address.  
35. Login as admin user of harbor.  
36. Click username and select about.  
37. Check if there is a default ca certificate download link.  
38. Login as user A.  
39. Click username and select about.  
40. Check if there is a default ca certificate download link.  
41. On a Docker client,Run `docker login <harbor_host>` command to login.  
42. Put the certificate file into /etc/docker/cert.d/VM_IP_address(or FQDN)/.  
43. Run `docker login <harbor_host>` command to login.

# Expected Outcome:

* In step10, there should no download link.
* In step21, user A should not see the download link.
* In step26and41, user cannot login.
* In step28and43, user can login success.
* In step40, there should no download link.

# Possible Problems:
None:
