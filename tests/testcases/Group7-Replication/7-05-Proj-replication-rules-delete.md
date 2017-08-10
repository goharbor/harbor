Test 7-05 Project replication rules delete
=======

# Purpose:

To verify that an admin user can delete replication rules.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.  

# Test Steps:

1. Login as admin user.  
2. In project replication page,delete an enabled rule.
3. Disable an enabled rule with running job, delete the rule while there is unfinished job.  
4. In project replication page, delete a disabled rule.

# Expected Outcome:

* In step2 rule can not be deleted.
* In step3 rule can not be deleted.  
* In step4 rule can be deleted.

# Possible Problems:
None
