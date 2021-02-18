Test 9-02 User Push unsigned images(DB mode)
=======

# Purpose:

To verify UI will difference unsigned images from signed images.

# References:
User guide

# Environment:

* This test requires one Harbor instance is running and available.
* A Linux host with Docker CLI (Docker client) installed.

# Test Steps:

1. Login UI and create a project.
2. Unset DOCKER_CONTENT_TRUST on Docker client and login Harbor.
3. Push a image to project created in step1.

# Expected Outcome:

* A red cross will displayed under signed column in UI.

# Possible Problems:
None
