Test 8-05 -Update email settings
=======

# Purpose:

To verify that an admin user can update email settings.

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* A linux host with Docker CLI installed (Docker client).

# Test Steps:

1. Login UI as admin user.
2. In configuration email page, change email settings.
3. Save settings and logout.
4. Login to check if email settings has been saved.
5. Click test mail server.

# Expected Outcome:

* In step4, email settings can be saved.
* In step5, if email settings are correct, test will successful.

# Possible Problems:
None
