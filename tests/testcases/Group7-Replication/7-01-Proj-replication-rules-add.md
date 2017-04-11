Test 7-01 - Create Replication Policy
=======

# Purpose:

To verify that admin user can create a replication rule.

# References:

User guide

# Environment:

* This test requires that at least two Harbor instances are running and available.

# Test Steps:

1. Login UI as admin user.
2. In Project replication page, add a replication rule.
3. Add another rule with different name use the same endpoint.  

# Expected Outcome:

* In step2, a rule with given name will be added.  
* In step3, rule add will fail.  

# Possible Problems:
None
