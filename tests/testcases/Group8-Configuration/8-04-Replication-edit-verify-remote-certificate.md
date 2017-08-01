Test 8-04 - Update-verify-remote-cert-settings
=======

# Purpose:

To verify that an admin user can update verify remote cert setting.  

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. Login UI as admin user.
2. In configuration replication page, uncheck verify remote certificate settings.
3. Save settings.
4. Add an end point that use a selfsigned certificate.
5. Add a replication rule of any project and enable the rule.  

# Expected Outcome:

* In step5, project can be replicated. 

# Possible Problems:
None
