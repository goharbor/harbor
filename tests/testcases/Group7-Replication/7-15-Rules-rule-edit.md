Test 7-15 Rules: rule edit
=======

# Purpose:

To verify admin user can edit replication rules.

# References:
User guide

# Environment:

* This test requires at least two Harbor instances are running and available.
* Project has at least one enabled and one disabled replication rule.  

# Test Steps:

1. Login UI as admin user.  
2. In replication page,edit name, description, enable status, endpoint info of enabled replication rules.
3. Push an image to the project.  
4. In replication page,edit name, description, enable status, endpoint info of disabled replication rules.

# Expected Outcome:

* In step2, rules can not be edited.  
* In step4, rules can be edited. If enable rule, replication job will start.   

# Possible Problems:
None
