Test 9-03 User pull signed images(DB mode)
=======

# Purpose:

To verify user can pull signed images.

# References:
User guide

# Environment:

* This test requires one Harbor instance is running and available.
* A Linux machine with Docker CLI(Docker client) installed.

# Test Steps:
**NOTE:**
In below test, project X should be replaced by an existing project and <harbor_ip> should be replaced by your harbor's ip or FQDN. If you are using a self-signed certificate,make sure to copy the CA root cert into ```/etc/docker/certs.d/<harbor_ip>``` and ```$HOME/.docker/tls/<harbor_ip>:4443/```

1. Login UI.
2. On Docker client, run
```sh
export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://<harbor_ip>:4443
```
and login Harobr.
3. Pull an image from project X.

# Expected Outcome:

* Image can be pulled successful.

# Possible Problems:
None
