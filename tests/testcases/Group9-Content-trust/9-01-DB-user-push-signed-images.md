Test 9-01 User push signed images(DB mode)
=======

# Purpose:

To verify user can push images with content trust enabled.

# References:
User guide

# Environment:

* This test requires one Harbor instance is running and available.  
* A Linux host with Docker CLI installed (Docker client).  

# Test Steps:
**NOTE:**  
In below test, <harbor_ip> should be replaced by your harbor's ip or FQDN. If you are using a self-signed certificate,make sure to copy the CA root cert into ```/etc/docker/certs.d/<harbor_ip>``` and ```$HOME/.docker/tls/<harbor_ip>:4443/```  

1. Login UI and create a project.  
2. On Docker client, run  
```sh
export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://<harbor_ip>:4443
```
and login Harbor.  
3. Push an image to the project created in step1.  


# Expected Outcome:

* In step3, Docker client will sign and push the image, a green tick will show in UI.  

# Possible Problems:
None
