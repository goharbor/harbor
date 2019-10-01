Test 7-02 - Edit replication rules
=======

# Purpose:

To verify admin user can edit repliciation rules.

# References:
User guide

# Environment:

* This test requires that at least two Harbor instances are running and available.
* Need at least one replication rule has been created.

# Test Steps:

1. Login UI as admin user.
2. In `Administration->Replications` page, choose a rule and edit its configurations.

Repeat steps 1-2 under `Projects->Project_Name->Replication` page.

# Expect Outcome:

* In step 2, Rule can be edited.

# Possible Problems:
None
