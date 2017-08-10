Test 7-04 Project replication rules enable
=======

# Purpose:

To verify an admin user can enable disabled rules.

# References:
User guide

# Environment:

* This test requires that at least two Harbor instance are running and avaliable.  
* Need at least one project that has at least one disabled rule.

# Test Steps:

1. Login UI as admin user.  
2. Push an image to project.  
3. In project replication page, enable a disabled rule.

# Expected Outcome:

* In step3 rule become enabled, replication should start.   

# Possible Problems:
None
