Test 7-21 - Replication-push-images  
=======

# Purpose:

To verify pushed images can be replicated to remote.  

# Reference:

User guide

# Environment:

* This test requires that at least two Harbor instance are running and available.  
* Need at least one project has at least one replication rule. 

# Test Steps:  
**NOTE:** In below test, Harbor instance should have at least one available endpoint.  

1. Login UI as admin user;  
2. Create a project and add an enabled replication rule.  
3. Push an image to the created project.  
4. Check replication job.  
5. Check remote registry to see if the image has been replicated if project existing on remote registry.  
6. Check remote registry to see if the image has been replicated if project not existing on remote registry.  

# Expect Outcome:  

* On remote registry can see duplicated projects and images. 

# Possible Problems:
None
