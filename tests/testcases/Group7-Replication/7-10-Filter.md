Test 7-10 - Filter
=======

# Purpose:

To verify the filer in replication rule can work as expected.

# References:
User guide

# Environment:

* This test requires that at least two Harbor instances are running and available.
* Create a new replication rule whose triggering condition is set to manual and tag filer is set to release-\*.

# Test Steps:

1. Push an image whose tag is release-1.0.
2. Push an image whose tag is dev-1.0.
3. Login UI with admin user.
4. Select the rule and click `REPLICATE` button.

# Expect Outcome:

* In step 4, one or two jobs should be started. The image in step 1 should be replicated to the remote registry and the image in step 2 should not be replicated to remote registry.

# Possible Problems:
None
