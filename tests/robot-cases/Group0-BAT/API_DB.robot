*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/APITest-Util.robot
Resource  ../../resources/Docker-Util.robot
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
    [tags]  gc
    Harbor API Test  ./tests/apitests/python/test_garbage_collection.py

Test Case - Add Private Project Member and Check User Can See It
    [tags]  pro_private_member
    Harbor API Test  ./tests/apitests/python/test_add_member_to_private_project.py

Test Case - Delete a Repository of a Certain Project Created by Normal User
    [tags]  del_repo
    Harbor API Test  ./tests/apitests/python/test_del_repo.py

Test Case - Add a System Global Label to a Certain Tag
    [tags]  global_lbl
    Harbor API Test  ./tests/apitests/python/test_add_sys_label_to_tag.py

Test Case - Add Replication Rule
    [tags]  replication_rule
    Harbor API Test  ./tests/apitests/python/test_add_replication_rule.py

Test Case - Edit Project Creation
    [tags]  pro_creation
    Harbor API Test  ./tests/apitests/python/test_edit_project_creation.py

Test Case - Scan Image
    [tags]  scan
    Harbor API Test  ./tests/apitests/python/test_scan_image.py

Test Case - Manage Project Member
    [tags]  pro_member
    Harbor API Test  ./tests/apitests/python/test_manage_project_member.py

Test Case - Project Level Policy Content Trust
    [tags]  content_trust
    Harbor API Test  ./tests/apitests/python/test_project_level_policy_content_trust.py

Test Case - User View Logs
    [tags]  logs
    Harbor API Test  ./tests/apitests/python/test_user_view_logs.py

Test Case - Scan All Images
    [tags]  scan_all
    Harbor API Test  ./tests/apitests/python/test_scan_all_images.py

Test Case - List Helm Charts
    [tags]  list_helm_charts
    Harbor API Test  ./tests/apitests/python/test_list_helm_charts.py

Test Case - Assign Sys Admin
    [tags]  sys_admin
    Harbor API Test  ./tests/apitests/python/test_assign_sys_admin.py

Test Case - Retag Image
    [tags]  retag
    Harbor API Test  ./tests/apitests/python/test_retag.py

Test Case - Robot Account
    [tags]  robot_account
    Harbor API Test  ./tests/apitests/python/test_robot_account.py

Test Case - Sign A Image
    [tags]  sign
    Harbor API Test  ./tests/apitests/python/test_sign_image.py

Test Case - Project Quota
    [tags]  quota
    Harbor API Test  ./tests/apitests/python/test_project_quota.py

Test Case - System Level CVE Whitelist
    [tags]  sys_cve
    Harbor API Test  ./tests/apitests/python/test_sys_cve_whitelists.py

Test Case - Project Level CVE Whitelist
    [tags]  pro_cve
    Harbor API Test  ./tests/apitests/python/test_project_level_cve_whitelist.py

Test Case - Tag Retention
    [tags]  tag_retention
    Harbor API Test  ./tests/apitests/python/test_retention.py

Test Case - Health Check
    [tags]  health_check
    Harbor API Test  ./tests/apitests/python/test_health_check.py

