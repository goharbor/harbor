Test 7-22 - Replication-delete-images  
=======

# Purpose:

To verify remote images can be delete when it is deleted.  

# Reference:

User guide  

# Environment:

* This test requires that at least two Harbor instance are running and available.
* Need at least one project has at least one replication rule.

# Test Steps:  
**NOTE:** In below test, Harbor instance should have at least one available endpoint.  

1. Login source registry UI as admin user.  
2. Create a project and add an enabled replication rule.
3. Push at least one image to the created project, and wait until replication job is done.  
4. Check the job log.  
5. Delete a pushed image in UI from the source registry.  
6. Check the replication job.  
7. Check remote registry to see if the image has been deleted.  

# Expeced Outcome:

* In step3, the remote will see replicated project and images.  
* In step6, there will be a replication job and the deleted image will be deleted from remote too.  
* In step7, images on remote registry should be deleted.  

# Possible Problems:
None
