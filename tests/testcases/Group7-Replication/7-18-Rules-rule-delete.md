Test 7-18 Rules rule delete
=======

# Purpose:

To verify admin user can delete replication rules.

# References:
User guide

# Environment:

* This test requires at leaset one Harbor instance is running and available.
* At least exist one project.
* Projects has at least one replication rule or more.

# Test Steps:

1. Login UI as admin user.  
2. In replication page, delete an enabled rule.  
3. In replication page, delete a disabled rule.  

# Expected Outcome:

* In step2, rule cannot be deleted.
* In step3, rule can be deleted.

# Possible Problems:
None
