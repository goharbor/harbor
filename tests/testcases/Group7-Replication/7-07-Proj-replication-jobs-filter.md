Test 7-07 Project replication jobs filter
=======

# Purpose:

To verify replication jobs filter works correctly.  

# References:
User guide

# Environment:

* This test requires at least two Harbor instance are running and available, and a project with a few jobs.

# Test Steps:

1. Login source registry as admin user.
2. Create some project and create replication rules.
3. Push some images into the project to start replication jobs and keep them running for step 4.  
4. In project replication page, input some character in jobs log filter, and then clear the filter.  

# Expected Outcome:

* In step4, jobs can be filtered, and after clear filter, all jobs are shown again.   

# Possible Problems:
None
