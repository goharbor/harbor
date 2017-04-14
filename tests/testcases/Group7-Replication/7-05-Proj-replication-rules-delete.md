Test 7-05 Project replication rules delete
=======

# Purpose:

To verify that an admin user can delete replication rules.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and avaiable.  

# Test Steps:

1. Login as admin user.  
2. In project replication page,delete an enabled rule.
3. In project replication page, delete a disabled rule.

# Expected Outcome:

* In step2 rule cannot be deleted.
* In step3 rule can be deleted.

# Possible Problems:
None
