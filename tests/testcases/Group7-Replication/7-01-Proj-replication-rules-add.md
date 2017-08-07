Test 7-01 - Create Replication Policy
=======

# Purpose:

To verify that admin user can create a replication rule.

# References:

User guide

# Environment:

* This test requires that at least two Harbor instances are running and available.

# Test Steps:

1. Login UI as admin user.
2. In Project replication page, add a replication rule using an existing endpoint with enable checked.  
3. In Project replication page, add a replication rule using an existing endpoint without enable checked.  
4. In Project replication page, add a replication rule using a new endpoint.
5. In Project replication page, add a replication rule using a new endpoint. Provide invalid values of input to see if validation works:

* endpoint name or ip address duplicate with an existing endpoint.  
* endpoint ip address incorrect.  
* endpoint username or password incorrect.  

6. Add another rule with different name using the same endpoint.  


# Expected Outcome:

* In step2, a rule with given name will be added. And all images will be replicated to remote.   
* In step3, a rule will be added and enabled,no image replication job is started.  
* In step4, a rule using new endpoint will be added.  
* In step5, rule can not be added if use duplicate name or ip, 
* In step5, if input wrong username/password/ip,rule can be added,but will cause test connection fail.  
* In step6, rule can not be added.  

# Possible Problems:
None
