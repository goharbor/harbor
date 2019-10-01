Test 11-03 - Add/remove labels to/from images
=======

# Purpose:

To verify that the users whose role >= project developer can add/remove labels to/from images.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* At least one global level label and one project level label are created.

# Test Steps:

1. The project developer user logs in to the UI.
2. The user add a global level label to an image from the UI.
3. The user add a project level label to an image from the UI.
4. The user removes the global label from the image.
5. The user removes the project label from the image.
6. Login in to the UI as a project admin user and repeat the Step2-5.
7. Login in to the UI as a system admin user and repeat the Step2-5.
8. Login in to the UI as a project guest user and try to add a label to an image.

# Expected Outcome:

* In Step2, the global level label can be added to the image successfully.
* In Step3, the project level label can be added to the image successfully.
* In Step4, the global level label can be removed from the image successfully.
* In Step5, the project level label can be removed from the image successfully.
* In Step6, the project admin user can do the same operations as the project developer user.
* In Step7, the system admin user can do the same operations as the project developer user.
* In Step8, the project guest user can not add a label to an image.

# Possible Problems:
None