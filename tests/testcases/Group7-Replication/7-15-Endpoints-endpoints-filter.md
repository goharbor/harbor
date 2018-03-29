Test 7-15 Endpoints endpoints filter
=======

# Purpose:

To verify endpoints filter works correctly.

# References:
User guide

# Environment:

* This test requires at least one Harbor instance is running and available. There are a few endpoints in the system.  

# Test Steps:

1. Login UI as admin user.    
2. In `Administration->Registries`,add some endpoints.    
3. Input some characters in endpoints filter and then clear the filter.  

# Expected Outcome:

* In step3, endpoints can be filtered, after clearing the filter, all endpoints should be shown again.  

# Possible Problems:

None
