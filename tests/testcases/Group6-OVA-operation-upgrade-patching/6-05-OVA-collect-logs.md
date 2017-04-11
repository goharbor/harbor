Test 6-05 - OVA Collect Logs
=======

# Purpose:

To verify that the logs of an OVA version of Harbor can be retrieved.

# References:
User guide, installation guide of Harbor OVA version.

# Environment:

* A running version of Harbor OVA version.
* A vCenter, at least an ESX host.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. From vSphere Web Client, import Harbor's OVA onto an ESX host.
2. In the deployment wizard, enter different passwords of Linux root user, Harbor admin user and MySQL root user.
3. In the deployment wizard,enable root login 
4. Power on the imported OVA.
5. Wait a few minutes for the VM's booting and its IP address comes up in vCenter. (may need to refresh in Web Client)
6. Open console from vSphere client and login root or login from ssh
7. Run log collect script /harbor/script/collect.sh

# Expected Outcome:

* An archive file named harbor_logs.tar.gz can be created and the archive contains correct log files

# Possible Problems:
None
