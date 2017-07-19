Test 7-23- Replication verify remote cert  
=======

# Purpose:  
To verify that if the verify remote cert setting works.  

# Reference:

User guide  

# Environment:

* This test requires at least two Harbor instances run and available.  
* The remote endpoint instance should be configured to use https and the remote harbor instance is set up with self-signed certificate.  

# Test Steps:  

1. Login source registry as admin user.  
2. In configuration page, make sure replication verify remote cert is checked on(by default).  
3. In replication page, add an https endpoint use self-signed cert. Click test connection button before save.  
4. In configuraiton page, uncheck replication verify remote cert checkbox and save.  
5. In replication page, choose the added https endpoint, and click test connection.  

# Expected Outcome:

* In step3, test will fail with a certificate error.  
* In step5, the test will successful.  

# Possible Problems:

None
