Test 7-11 Endpoints endpoints add
=======

# Purpose

To verify admin user can add an endpoint

# References:

User guide

# Environments:

* This test requires at least two Harbor instance is running and available.

# Test Steps:

1. Login UI as admin user.
2. In replication page, add an endpoint with valid ip/hostname/username/password,click test connection to test connection.  
3. In replication page, add an endpoint with invalid ip/hostname/username/password, click test connection to test connection.  

# Expected Outcome:

* In step2 if endpoint is alive,test connection will successful, endpoint will add success.
* In step3 test connection will fail, endpoint will add success.  
# Possible Problem:
None
