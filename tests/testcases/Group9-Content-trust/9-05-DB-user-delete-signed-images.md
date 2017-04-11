Test 9-05 User delete signed images(DB mode)
=======

# Purpose:

To verify whether user can delete signed images.

# References:
User guide

# Environment:

* This test requires one Harbor instance is running and avialable.  
* A Linux host with Docker CLI(Docker client) installed.  

# Test Steps:

1. Login UI and create a project.  
2. On a Docker client,run   
```sh 
export DOCKER_CONTENT_TRUST=1   
export DOCKER_CONTENT_TRUST_SERVER=https://<harbor_ip>:4443  
```
and login Harbor.  
3. Push an image to project created in step1.  
4. Delete the pushed image.  
5. Delete notary tag according to message from UI in step4.  
6. Delete the image again.

# Expected Outcome:

* In step4, image cannot be deleted.  
* In step6, image can be deleted.

# Possible Problems:
None
