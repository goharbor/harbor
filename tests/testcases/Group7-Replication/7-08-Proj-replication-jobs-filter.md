Test 7-08 Project replication jobs filter
=======

# Purpose:

To verify replication jobs filter works correctly.  

# References:
User guide

# Environment:

* This test requires at least two Harbor instance are running and available, and a project with a few jobs.

# Test Steps:

1. Login as admin user.
2. In `Administration->Replications` page, click a rule which has replication jobs.
3. Input some characters in jobs log filter, and then clear the filter.  

Repeat steps 1-3 under `Projects->Project_Name->Replication` page.

# Expected Outcome:

* In step3, jobs can be filtered, and after clear filter, all jobs are shown again.   

# Possible Problems:
None
