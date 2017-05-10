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
3. Push an image to created project.  
4. Check replication job.  

# Expect Outcome:  

* On endpoint can see duplicated projects and images. 

# Possible Problems:
None
