Test 7-06 Project replication rules filter
=======

# Purpose:

To verify rule filter works correctly.  

# References:
User guide

# Environment:
* This test requires that two Harbor instance are running and avaiable and there are at least 5 rules of a project.  

# Test Steps:

1. Login source registry ui as user.
2. In project replication page,input some character in rule filter and then clear the filter.

# Expected Outcome:

* In step2, rules can be filtered, after clearing filter, all rules are shown again.

# Possible Problems:
None
