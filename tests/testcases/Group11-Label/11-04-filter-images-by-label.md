Test 11-04 - Filter images by label
=======

# Purpose:

To verify that the images can be filtered by labels.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Create at least two labels and add one of them to an image.

# Test Steps:

1. The project guest user logs in to the UI.
2. The user filters the images by the label that has been added to the image.
3. The user filters the images by the label that has not been added to the image.

# Expected Outcome:

* In Step2, the image list contains the image which is labeled.
* In Step3, the image list doesn't contain the image which is labeled.

# Possible Problems:
None