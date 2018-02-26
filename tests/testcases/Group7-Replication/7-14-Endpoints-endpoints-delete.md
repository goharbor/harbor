Test 7-14 Endpoints endpoint delete
=======

# Purpose

To verify admin user can delete an endpoint.

# References:
User guide

# Environments:

* This test requires one Harbor instance is running and available.
* At least one endpoint should exist.

# Test Steps:

1. Login UI as admin user.  
2. In `Administration->Registries` page, delete an endpoint in use by a rule.  
3. In `Administration->Registries` page, delete an endpoint not in use by a rule.  

# Expected Outcome:

* In step2, endpoint can not be deleted.  
* In step3, endpoint can be deleted.  

# Possible Problems:
None
