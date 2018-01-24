Test 7-03 Delete replication rules  
=======

# Purpose:

To verify that an admin user can delete replication rules.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.  

# Test Steps:

1. Login as admin user.  
2. In `Administration->Replications` page,delete a rule which has no pending/running/retrying jobs.
3. Delete a rule which has pending/running/retrying jobs.  

Repeat steps 1-3 under `Projects->Project_Name->Replication` page.

# Expected Outcome:

* In step2 rule can be deleted.
* In step3 rule can not be deleted.  

# Possible Problems:
None
