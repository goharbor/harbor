Test 11-02 - Project admins manage project level labels
=======

# Purpose:

To verify that the project administrators can manage(CURD) the project level labels.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.

# Test Steps:

1. The project admin user logs in to the UI.
2. The project admin user creates a project level label under a project from the UI.
3. The project admin user edits the label created in Step1.
4. The project admin user deletes the label created in Step1.
5. The project developer user logs in to the UI.
6. The project developer user tries to create a project level label under a project from the UI.

# Expected Outcome:

* In Step2, the label can be created successfully.
* In Step3, the label can be updated successfully.
* In Step4, the label can be deleted successfully.
* In Step6, the project developer user can not create project level labels.

# Possible Problems:
None