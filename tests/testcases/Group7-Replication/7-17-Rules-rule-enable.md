Test 7-17 Rules rule enable
=======

# Purpose:

To verify admin user can enable disabled rules.

# References:
User guide

# Environment:

* This test requires two Harbor instances are running and available.
* Projects has at least one disabled replication rule or more.  

# Test Steps:

1. Login UI as admin user.  
2. Push an image to a project has at least one disabled rule.  
3. In replication page, choose a disabled rule,enable it.

# Expected Outcome:

* Disabled rule can be enabled,after enable the rule, replication job will start.  

# Possible Problems:
None
