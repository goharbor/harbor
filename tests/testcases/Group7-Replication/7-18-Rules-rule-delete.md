Test 7-18 Rules rule delete
=======

# Purpose:

To verify admin user can delete replication rules.

# References:
User guide

# Environment:

* This test requires at least two Harbor instances are running and available.
* Source registry has at least one endpoint.  
* Projects has at least one replication rule or more.

# Test Steps:

1. Login UI as admin user.  
2. In replication page, delete an enabled rule.  
3. Disable an enabled rule with running job, delete the rule while there is unfinished job.  
4. In replication page, delete a disabled rule.  

# Expected Outcome:

* In step2, rule can not be deleted.
* In step3, rule can not be deleted.
* In step4, rule can be deleted.  

# Possible Problems:
None
