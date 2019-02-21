Test 7-04 Filter replication rules  
=======

# Purpose:

To verify rule filter works correctly.  

# References:
User guide

# Environment:
* This test requires that two Harbor instance are running and available.  

# Test Steps:

1. Login as admin user.
2. In `Administration->Replications`,input some character in rule filter and then clear the filter.

Repeat steps 1-2 under `Projects->Project_Name->Replication` page.

# Expected Outcome:

* In step2, rules can be filtered, after clearing filter, all rules are shown again.

# Possible Problems:
None
