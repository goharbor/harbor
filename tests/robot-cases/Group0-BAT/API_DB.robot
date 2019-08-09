*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/APITest-Util.robot
Resource  ../../resources/Docker-Util.robot
Library  ../../apitests/python/library/Harbor.py  ${SERVER_CONFIG}
Library  OperatingSystem
Library  String
Library  Collections
Library  requests
Library  Process
Default Tags  APIDB

*** Variables ***
${SERVER}  ${ip}
${SERVER_URL}  https://${SERVER}
${SERVER_API_ENDPOINT}  ${SERVER_URL}/api
&{SERVER_CONFIG}  endpoint=${SERVER_API_ENDPOINT}  verify_ssl=False

*** Test Cases ***
Test Case - Garbage Collection
    Harbor API Test  ./tests/apitests/python/test_garbage_collection.py
Test Case - Add Private Project Member and Check User Can See It
    Harbor API Test  ./tests/apitests/python/test_add_member_to_private_project.py
Test Case - Delete a Repository of a Certain Project Created by Normal User
    Harbor API Test  ./tests/apitests/python/test_del_repo.py
Test Case - Add a System Global Label to a Certain Tag
    Harbor API Test  ./tests/apitests/python/test_add_sys_label_to_tag.py
# Test Case - Add Replication Rule
#    Harbor API Test  ./tests/apitests/python/test_add_replication_rule.py
Test Case - Edit Project Creation
    Harbor API Test  ./tests/apitests/python/test_edit_project_creation.py
Test Case - Scan Image
    Harbor API Test  ./tests/apitests/python/test_scan_image.py
Test Case - Manage Project Member
    Harbor API Test  ./tests/apitests/python/test_manage_project_member.py
Test Case - Project Level Policy Content Trust
    Harbor API Test  ./tests/apitests/python/test_project_level_policy_content_trust.py
Test Case - User View Logs
    Harbor API Test  ./tests/apitests/python/test_user_view_logs.py
Test Case - Scan All Images
    Harbor API Test  ./tests/apitests/python/test_scan_all_images.py
Test Case - List Helm Charts
    Harbor API Test  ./tests/apitests/python/test_list_helm_charts.py
Test Case - Assign Sys Admin
    Harbor API Test  ./tests/apitests/python/test_assign_sys_admin.py
Test Case - Retag Image
    Harbor API Test  ./tests/apitests/python/test_retag.py
Test Case - Robot Account
    Harbor API Test  ./tests/apitests/python/test_robot_account.py
Test Case - Sign A Image
    Harbor API Test  ./tests/apitests/python/test_sign_image.py
