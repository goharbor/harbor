Test 7-09 Project replication jobs log view
=======

# Purpose:

To verify admin user can view replication jobs log.

# References:
User guide

# Environment:

* This test requires at least two Harbor instance are running and available.
* At least one replication job exist.

# Test Steps:

1. Login source registry UI as admin user.
2. In `Administration->Replications` page, select a replication job and view job log.

Repeat steps 1-2 under `Projects->Project_Name->Replication` page.

# Expected Outcome:

* In step2, user can view job log.

# Possible Problems:
None
