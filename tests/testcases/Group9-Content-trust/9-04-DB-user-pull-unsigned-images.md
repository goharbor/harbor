Test 9-04 User pull unsigned images(DB mode)
=======

# Purpose:

To verify whether user can pull unsigned images with content trust enabled.

# References:
User guide

# Environment:

* This test requires one Harbor instance is running and available.
* Harbor is set to authenticate against a local database.The user data is stored in a local database.  
* A Linux host with Docker CLI(Docker client) installed.

# Test Steps:

1. Login UI and create a project.
2. On Docker client, unset DOCKER_CONTENT_TRUST and login Harbor.  
3. Push an image to the project created in step1.  
4. Reset DOCKER_CONTENT_TRUST to 1.
5. Pull the unsigned image.

# Expected Outcome:

* User cannot pull unsigned images with content trust enabled.

# Possible Problems:
None
