Test 7-12 Endpoints endpoints add
=======

# Purpose

To verify admin user can add an endpoint

# References:

User guide

# Environments:

* This test requires at least two Harbor instance is running and available.

# Test Steps:

1. Login UI as admin user.
2. In `Administration->Registries` page, add an endpoint with valid URL(HTTP), username and password, click test connection to test connection and save the endpoint.  
3. In `Administration->Registries` page, add an endpoint with invalid URL(HTTP), username or password, click test connection to test connection and save the endpoint.  
4. In `Administration->Registries` page, add an endpoint with valid URL(HTTPS with self-signed certificate), username and password, select the `Verify Remote Cert` checkbox and click test connection to test connection and save the endpoint.  
5. In `Administration->Registries` page, add an endpoint with valid URL(HTTPS with self-signed certificate), username and password, uncheck the `Verify Remote Cert` checkbox and click test connection to test connection and save the endpoint.  

# Expected Outcome:

* In step2 if endpoint is alive,test connection will successful, endpoint will add success.
* In step3 test connection will fail, endpoint will add success.  
* In step4 test connection will fail, endpoint will add success.
* In step5 if endpoint is alive,test connection will successful, endpoint will add success.
# Possible Problem:
None
