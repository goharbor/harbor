Test 9-03 User pull signed images(DB mode)
=======

# Purpose:

To verify user can pull signed images.

# References:
User guide

# Environment:

* This test requires one Harbor instance is running and available.
* A Linux machine with Docker CLI(Docker client) installed.

# Test Steps:
NOTE:  
In below test, project X should be replaced by an existing project.

1. Login UI.
2. On Docker client, follow [Set up notary](../../../../docs/use_notary.md) to set up notary and login Harobr.  
3. Pull an image from project X.  

# Expected Outcome:

* Image can be pulled successful.

# Possible Problems:
None
