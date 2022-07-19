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
    [Tags]  gc
    Harbor API Test  ./tests/apitests/python/test_garbage_collection.py

Test Case - Add Private Project Member and Check User Can See It
    [Tags]  private_member
    Harbor API Test  ./tests/apitests/python/test_add_member_to_private_project.py

Test Case - Delete a Repository of a Certain Project Created by Normal User
    [Tags]  del_repo
    Harbor API Test  ./tests/apitests/python/test_del_repo.py

Test Case - Add a System Global Label to a Certain Tag
    [Tags]  global_lbl
    Harbor API Test  ./tests/apitests/python/test_add_sys_label_to_tag.py

Test Case - Add Replication Rule
    [Tags]  replic_rule
    Harbor API Test  ./tests/apitests/python/test_add_replication_rule.py

Test Case - Edit Project Creation
    [Tags]  pro_creation
    Harbor API Test  ./tests/apitests/python/test_edit_project_creation.py

Test Case - Manage Project Member
    [Tags]  member
    Harbor API Test  ./tests/apitests/python/test_manage_project_member.py

Test Case - User View Logs
    [Tags]  view_logs
    Harbor API Test  ./tests/apitests/python/test_user_view_logs.py

Test Case - List Helm Charts
    [Tags]  list_helm_charts
    Harbor API Test  ./tests/apitests/python/test_list_helm_charts.py

Test Case - Assign Sys Admin
    [Tags]  assign_adin
    Harbor API Test  ./tests/apitests/python/test_assign_sys_admin.py

Test Case - Copy Artifact Outside Project
    [Tags]  copy_artifact
    Harbor API Test  ./tests/apitests/python/test_copy_artifact_outside_project.py

Test Case - Robot Account
    [Tags]  robot_account
    Harbor API Test  ./tests/apitests/python/test_robot_account.py

Test Case - Sign A Image
    [Tags]  sign_image
    Harbor API Test  ./tests/apitests/python/test_sign_image.py

Test Case - Project Quota
    [Tags]  quota
    Harbor API Test  ./tests/apitests/python/test_project_quota.py

Test Case - System Level CVE Allowlist
    [Tags]  sys_cve
    Harbor API Test  ./tests/apitests/python/test_sys_cve_allowlists.py

Test Case - Project Level CVE Allowlist
    [Tags]  pro_cve
    Harbor API Test  ./tests/apitests/python/test_project_level_cve_allowlist.py

Test Case - Tag Retention
    [Tags]  tag_retention
    Harbor API Test  ./tests/apitests/python/test_retention.py

Test Case - Health Check
    [Tags]  health
    Harbor API Test  ./tests/apitests/python/test_health_check.py

Test Case - Push Index By Docker Manifest
    [Tags]  push_index
    Harbor API Test  ./tests/apitests/python/test_push_index_by_docker_manifest.py

Test Case - Push Chart By Helm3 Chart CLI
    [Tags]  push_chart
    Harbor API Test  ./tests/apitests/python/test_push_chart_by_helm3_chart_cli.py

Test Case - Push Chart By Helm3.7 Chart CLI
    [Tags]  push_chart_by_Helm3.7
    Harbor API Test  ./tests/apitests/python/test_push_chart_by_helm3.7_chart_cli.py

Test Case - Push Cnab Bundle
    [Tags]  push_cnab
    Harbor API Test  ./tests/apitests/python/test_push_cnab_bundle.py

Test Case - Tag CRUD
    [Tags]  tag_crud
    Harbor API Test  ./tests/apitests/python/test_tag_crud.py

Test Case - Scan Image
    [Tags]  scan
    Harbor API Test  ./tests/apitests/python/test_scan_image_artifact.py

Test Case - Scan Image In Public Project
    [Tags]  scan_public_project
    Harbor API Test  ./tests/apitests/python/test_scan_image_artifact_in_public_project.py

Test Case - Scan All Images
    [Tags]  scan_all
    Harbor API Test  ./tests/apitests/python/test_system_level_scan_all.py

Test Case - Stop Scan Image
    [Tags]  stop_scan
    Harbor API Test  ./tests/apitests/python/test_stop_scan_image_artifact.py

Test Case - Stop Scan All Images
    [Tags]  stop_scan_all
    Harbor API Test  ./tests/apitests/python/test_system_level_stop_scan_all.py

Test Case - Registry API
    [Tags]  reg_api
    Harbor API Test  ./tests/apitests/python/test_registry_api.py

Test Case - Push Image With Special Name
    [Tags]  special_repo_name
    Harbor API Test  ./tests/apitests/python/test_push_image_with_special_name.py

Test Case - Push Artifact With ORAS CLI
    [Tags]  oras
    Harbor API Test  ./tests/apitests/python/test_push_files_by_oras.py

Test Case - Push Singularity file With Singularity CLI
    [Tags]  singularity
    Harbor API Test  ./tests/apitests/python/test_push_sif_by_singularity.py

Test Case - Push Chart File To Chart Repository By Helm V2 With Robot Account
    [Tags]  helm2
    Harbor API Test  ./tests/apitests/python/test_push_chart_by_helm2_helm3_with_robot_Account.py

Test Case - Replication From Dockerhub
    [Tags]  replic_dockerhub
    Harbor API Test  ./tests/apitests/python/test_replication_from_dockerhub.py

Test Case - Proxy Cache
    [Tags]  proxy_cache
    Harbor API Test  ./tests/apitests/python/test_proxy_cache.py

Test Case - Tag Immutability
    [Tags]  tag_immutability
    Harbor API Test  ./tests/apitests/python/test_tag_immutability.py

Test Case - P2P
    [Tags]  p2p
    Harbor API Test  ./tests/apitests/python/test_p2p.py

Test Case - Metrics
    [Tags]  metrics
    Harbor API Test  ./tests/apitests/python/test_verify_metrics_enabled.py

Test Case - Project Level Policy Content Trust
    [Tags]  content_trust
    Harbor API Test  ./tests/apitests/python/test_project_level_policy_content_trust.py

Test Case - Webhook CRUD
    [Tags]  webhook
    Harbor API Test  ./tests/apitests/python/test_webhook_crud.py

Test Case - Cosign Sign Artifact
    [Tags]  cosign
    Harbor API Test  ./tests/apitests/python/test_cosign_sign_artifact.py