Test 7-05 - Manual trigger  
=======
  
# Purpose:  
  
To verify admin user can trigger a replication manually.  
  
# References:  
User guide  
  
# Environment:  

* This test requires that at least two Harbor instances are running and available.  
* Create a new replication rule whose triggering condition is set to manual and no filter is configured.  
* Need at least one image is pushed to the project that the rule is applied to
  
# Test Steps:  
  
1. Login UI as admin user.  
2. In `Administration->Replications` page, choose the rule and click the `REPLICATE` button.  

Repeat steps 1-2 under `Projects->Project_Name->Replication` page.

# Expect Outcome:  
  
* In step 2, the operation should success and a few jobs should be started. All the images under the project should be replicated to the remote registry.  
  
# Possible Problems:  
None  
