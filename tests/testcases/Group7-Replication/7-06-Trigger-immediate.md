Test 7-06 - Immediate trigger
=======

# Purpose:

To verify the immediate replication rule can work as expected.

# References:
User guide

# Environment:

* This test requires that at least two Harbor instances are running and available.
* Create a new replication rule whose triggering condition is set to immediate and no filter is configured.

# Test Steps:

1. Push an image to the project that the replication is applied to.
2. Login UI as admin user.
3. In `Administration->Replications` page, edit the rule and make sure the `Delete remote images when locally deleted` is checked.
4. Delete the image pushed in step 1 on UI.
5. In `Administration->Replications` page, edit the rule and make sure the `Delete remote images when locally deleted` is unchecked.
6. Push the image to the project that the replication is applied to again.
7. Delete the image pushed in step 6 on UI.

# Expect Outcome:

* In step 1, a job should be started and the image should be replicated to the remote registry.
* In step 4, a job should be started and the image should be deleted from the remote registry.
* In step 6, a job should be started and the image should be replicated to the remote registry.
* In step 7, a job should be started and the image should not be deleted from the remote registry.

# Possible Problems:
None
