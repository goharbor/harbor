Test 7-11 Project admin read-only privilege
=======

# Purpose

To verify project admin user has read-only privilege for replication rules and jobs.

# References:

User guide

# Environment:

* This test requires at least two Harbor instance are running and available.  
* At least a replication rule is created.  
* The replication rule has at least one job.  
* A member has been added to the project that the rule is applied to as admin.  

# Test Steps:

1. Login UI using the project admin user.  
2. Go to `Projects->Project_Name->Replication` page.  

# Expected Outcome:
 
* In step2, the user should have read-only privilege to all rules and jobs.  

# Possible Problems:
None
