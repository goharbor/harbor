Test 7-03 Project replication rules disable
=======

# Purpose:

To verify that an admin user can disable replication rules.  

# References:
User Guide

# Environment:

* This test requires that at least two Harbor instances are running and available.  
* Need at least one project that has at least one enabled rule.

# Test Steps:
1. Login UI as admin user.  
2. In project replication page, disable a rule.

# Expected outcome:

* Rule can be disabled. Replication jobs in queues will be canceled

# Possible Problems:
None
