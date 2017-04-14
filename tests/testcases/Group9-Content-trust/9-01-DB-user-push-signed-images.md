Test 9-01 User push signed images(DB mode)
=======

# Purpose:

To verify user can push images with content trust enabled.

# References:
User guide

# Environment:

* This test requires one Harbor instance is runnning and available.  
* A Linux host with Docker CLI installed (Docker client).  

# Test Steps:

1. Login UI and create a project.  
2. On Docker clinet, follow [Set up notary](../../../../docs/use_notary.md) to set up notary and login Harbor.  
3. Push an image to the project created in step1.  


# Expected Outcome:

* In step3, Docker client will sign and push the image, a green tick will show in UI.  

# Possible Problems:
None
