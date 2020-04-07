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
Test Case - Add Replication Rule
    Harbor API Test  ./tests/apitests/python/test_add_replication_rule.py
Test Case - Edit Project Creation
    Harbor API Test  ./tests/apitests/python/test_edit_project_creation.py
Test Case - Manage Project Member
    Harbor API Test  ./tests/apitests/python/test_manage_project_member.py
Test Case - Project Level Policy Content Trust
    Harbor API Test  ./tests/apitests/python/test_project_level_policy_content_trust.py
Test Case - User View Logs
    Harbor API Test  ./tests/apitests/python/test_user_view_logs.py
Test Case - List Helm Charts
    Harbor API Test  ./tests/apitests/python/test_list_helm_charts.py
Test Case - Assign Sys Admin
    Harbor API Test  ./tests/apitests/python/test_assign_sys_admin.py
Test Case - Copy Artifact Outside Project
    Harbor API Test  ./tests/apitests/python/test_copy_artifact_outside_project.py
Test Case - Robot Account
    Harbor API Test  ./tests/apitests/python/test_robot_account.py
Test Case - Sign A Image
    Harbor API Test  ./tests/apitests/python/test_sign_image.py
Test Case - Project Quota
   Harbor API Test  ./tests/apitests/python/test_project_quota.py
Test Case - System Level CVE Whitelist
    Harbor API Test  ./tests/apitests/python/test_sys_cve_whitelists.py
Test Case - Project Level CVE Whitelist
    Harbor API Test  ./tests/apitests/python/test_project_level_cve_whitelist.py
Test Case - Tag Retention
    Harbor API Test  ./tests/apitests/python/test_retention.py
Test Case - Health Check
    Harbor API Test  ./tests/apitests/python/test_health_check.py
Test Case - Push Index By Docker Manifest
    Harbor API Test  ./tests/apitests/python/test_push_index_by_docker_manifest.py
Test Case - Push Chart By Helm3 Chart CLI
    Harbor API Test  ./tests/apitests/python/test_push_chart_by_helm3_chart_cli.py
Test Case - Push Cnab Bundle
    Harbor API Test  ./tests/apitests/python/test_push_cnab_bundle.py
Test Case - Create/Delete tag
    Harbor API Test  ./tests/apitests/python/test_create_delete_tag.py
Test Case - Scan Image
    Harbor API Test  ./tests/apitests/python/test_scan_image_artifact.py
Test Case - Scan All Images
    Harbor API Test  ./tests/apitests/python/test_system_level_scan_all.py
Test Case - Registry API
    Harbor API Test  ./tests/apitests/python/test_registry_api.py
