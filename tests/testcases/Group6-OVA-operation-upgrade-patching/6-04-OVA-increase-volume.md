Test 6-04 - OVA Increase Volume Space
=======

# Purpose:

To verify that an OVA version of Harbor can be upgraded the volume to larger space.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:

* A running version of Harbor OVA version.
* A vCenter, at least an ESX host.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

NOTE:
* In this case /dev/sdc /data1_vg are examples, the disks and VG name should depends on the actual conditions.


1. From vSphere Web Client, import Harbor's OVA onto an ESX host.
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.
3. Power on the imported OVA.
4. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)
5. Open a browser and input http://VM_IP_address. 
6. Login as admin user of Harbor.
7. Click username and select about to check disk usage.	
8. On vSphere, open the console of Harbor's VM, log in as root user, type `df -h /data` to get disk space usage of /data volume.
9. Power off the VM.
10. Right click on the VM and select "Edit Settings".
11. Add one or more hard disks to VM.
12. Power on the VM.
13. On vSphere, open the console of Harbor's VM, log in as root user.
14. Run `pvcreate /dev/sdc [/dev/sdX(if there exists)]` to new added disks.
15. Run `vgextend data1_vg /dev/sdc` to extend data vg.
16. Run `lvresize -l +100%FREE /dev/data1_vg/data` to adjust size of logical volume.
17. Run `resize2fs /dev/data1_vg/data` to adjust size of file system.
18. If there are more disks, repeat step15-17.
19. Run `df -h /data` to check size of data volume.
20. Login harbor as admin.
21. Click username and select about to check the data volume size.

# Expected Outcome:

* Step7 should see the usage of disks.
* Step19 /data volume should be extended.
* Step21 should see disk space is extended.

# Possible Problems:
None
