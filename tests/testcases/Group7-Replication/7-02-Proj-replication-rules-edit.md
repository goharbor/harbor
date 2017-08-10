Test 7-02 - Project replication rules edit  
=======
  
# Purpose:  
  
To verify project manager can edit repliciation rules.  
  
# References:  
User guide  
  
# Environment:  

* This test requires that at least two Harbor instances are running and available.  
* Need at least one project that has at least one replication rule.  
  
# Test Steps:  
  
1. Login UI as admin user  
2. In project replication page, choose a disabled rule and edit rule name, description, rule enable status and endpoint name, username, password.  
3. In project replication page, choose an enabled rule and edit rule name, description, rule enable status and endpoint name, username, password.  
  
# Expect Outcome:  
  
* In step 2, Rule can be edited and error hint works correctly,if rule is enabled, should see jobs start.  
* In step 3, Rule can not be edited.  
  
# Possible Problems:  
None  
