Test 7-01 - Create Replication Policy
=======

# Purpose:

To verify that admin user can create a replication rule.

# References:

User guide

# Environment:

* This test requires that at least two Harbor instances are running and available.

# Test Steps:

1. Login UI as admin user.
2. In `Administration->Replications` page, create a new rule and fill in name and description.
3. Choose a project by clicking the icon on the right of `Projects`.
4. Add repository and tag filter.
5. Choose an endpoint, create a new one if no endpoint exists.
6. Select a triggering condition: Immediate/Manaul/Scheduled, check/uncheck the `Delete remote images when locally deleted` if choosing `Immediate`.
7. Check/uncheck the option `Replicate existing images immediately`.
8. Save the rule.

Repeat steps 1-8 under `Projects->Project_Name->Replication` page.

# Expected Outcome:

* In step8, a rule will be added.

# Possible Problems:
None
