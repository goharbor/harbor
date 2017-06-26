Test 7-15 Rules: rule edit
=======

# Purpose:

To verify admin user can edit replication rules.

# References:
User guide

# Environment:

* This test requires at least one Harbor instance is running and available.

# Test Steps:

1. Login UI as admin user.  
2. In replication page,edit name, description, enable status, endpoint info of enabled replication rules.
3. In replication page,edit name, description, enable status, endpoint info of disabled replication rules.

# Expected Outcome:

* In step2, rules can not be edited.  
* In step3, rules can be edited.  

# Possible Problems:
None
