Test 7-16 Rules rule disable
=======

# Purpose:

To verify admin user can disable replication rles.

# References:
User guide

# Environment:

* This test requires at least two Harbor instances are running and available.
* Need at least one project that has at least one enabled rule.  

# Test Steps:

1. Login UI as admin user.  
2. In Replication rule page, disable an enabled replication rule.

# Expected Outcome:

* Rule can be disabled and replication jobs in queues will be canceled. Jobs are running will be stopped.  

# Possible Problems:
None
