Test 7-12 Endpoints endpoints edit
=======

# Purpose

To verify admin user can edit edpoints.

# References:

User guide

# Environment:

* This test requires at least two Harbor instance are running and available.

# Test Steps:

1. Login UI as admin user.  
2. In replication page, choose an endpoint in use by a rule and edit setting.  
3. In replication page, choose an endpoint not in use by a rule and edit setting.

# Expected Outcome:

* In step2, endpoint info can not be edited.  
* In step3, endpoint info can be edited.  

# Possible Problems:
None
