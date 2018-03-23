Test 5-03 - OVA Garbage Collection
=======

# Purpose:

To verify that an OVA version of Harbor can perform garbage collection to release unused space of deleted images.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:
* This test requires an OVA binary of Harbor.
* A vCenter, at least an ESX host, and a network that supports DHCP.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:
1. Deploy an OVA version of Harbor, with "Garbage Collection" set to false.
2. Create a project in Harbor.
3. On a Docker client host, use `docker login <harbor_host>` command to log in as the admin user. 
4. Run some `docker push` to push some images to the project. The size of the images should be at least 500MB.
5. In Harbor's UI, delete the newly pushed images.
6. On vSphere, open the console of Harbor's VM, log in as root user, type `df -h /data` command to get the space usage of the /data volume. Take a note of the **Used** space.
7. Power off the VM.
8. Right click on the VM and select "Edit Settings".
9. Set "Garbage Collection" to true.
10. Power on the VM. 
11. Wait for a while until Harbor service is available (check by a browser)
12. On vSphere, open the console of Harbor's VM, log in as root user, type `df -h /data` command to get the space usage of the /data volume and compare with previous number.
13. Check the log file of garbage collection under /data to see if there was any error.
14. Repeat Step 3-7, to create some deleted images. 
15. Repeat Step 10-13, to see the garbage collection works on the second reboot.


# Expected Outcome:
* Step 12, the used space should be reduced. The space of deleted images should be recycled.
* Step 13, log file should contain no errors.
* Step 15, verify garbage collection works for the second time.

# Possible Problems:
None