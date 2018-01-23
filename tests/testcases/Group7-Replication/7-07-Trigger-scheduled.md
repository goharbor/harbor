Test 7-07 - Scheduled trigger  
=======
  
# Purpose:  
  
To verify the scheduled replication rule can work as expected.  
  
# References:  
User guide  
  
# Environment:  

* This test requires that at least two Harbor instances are running and available.  
* Create a new replication rule whose triggering condition is set to scheduled and no filter is configured.  
* Need at least one image is pushed to the project that the rule is applied to
  
# Test Steps:  
  
1. Login UI as admin user.  
2. Go to `Administration->Replications` page.  

# Expect Outcome:  
  
* In step 2, a few jobs should be started when the time configured in schedule trigger comes. All the images under the project should be replicated to the remote registry.  
  
# Possible Problems:  
None  
