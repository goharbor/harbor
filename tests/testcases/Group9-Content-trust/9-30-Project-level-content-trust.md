Test 9-30 Project enable content trust
=======

# Purpose:

To verify project level content trust works correctly.

# References:
User guide

# Environment:

* This test requires one Harbor instance is runnning and available.
* A Linux host with Docker CLI installed (Docker client).

# Test Steps:
**NOTE:**
In below test, <harbor_ip> should be replaced by your harbor's ip or FQDN. If you are using a self-signed certificate,make sure to copy the CA root cert into ```/etc/docker/certs.d/<harbor_ip>``` and ```$HOME/.docker/tls/<harbor_ip>:4443/```
project a should be replaced by meaingful and longer name.

1. Login UI and create a project a.
2. Push an image to project a.
3. On Docker clinet, run
```sh
export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTNET_TRUST_SERVER=https://<harbor_ip>:4443
```
and login Harbor.
4. Push an image to project a.
5. In project a configuration page, enabled project level content trust.
6. Pull the image the first time pushed.
7. Pull the image the second time pushed.

# Expected Outcome:

* In step6, the image can not be pulled.
* In step7, the image can be pulled.

# Possible Problems:
None
