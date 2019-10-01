Test 2-02 - User Push Multiple Images (DB Mode)
=======

# Purpose:

To verify that a non-admin user can push multiple images to a project when users are managed locally by Harbor (DB mode).

# References:
User guide

# Environment:
* This test requires that a Harbor instance is running and available.
* Harbor is set to authenticate against a local database. ( auth_mode is set to **db_auth** .) The user data is stored in a local database.
* A linux host with Docker CLI installed (Docker client).
* At least a non-admin user and the user has at least a project as project admin.

# Test Steps:

**NOTE:**
In below test, user A is non-admin user. User A and project X should be replaced by longer and meaningful names.

1. Log in to UI as user A (non-admin).
2. Verify User A has at least a project X with the role of project admin.
3. On a Docker client, log in as User A and run `docker push` to push an image with tag (e.g. nginx:1.5) to project X.
4. Continue to push at least 5 images with different tags, for example, nginx:1.6, nginx:1.7, nginx:1.8, nginx:release .
5. In the UI, go to "My Projects" to see if all images/tags are properly displayed.
6. Checks the detail info of each tag.
7. Enter keyword to search(filter) images/tags of a project.
8. On a Docker client, log in as User A and run `docker pull` to pull images with different tags from project X.

# Expected Outcome:
* Step 3-5, images can be pushed to project X and can be shown in UI.
* Step 6, image/tag info should be correct.
* Step 7, image/tag should be filtered and results should be matched the search criteria.
* Step 8, images can be pulled from project X.

# Possible Problems:
None