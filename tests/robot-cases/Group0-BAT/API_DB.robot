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
Test Case - Add Private Project Member and Check User Can See It
    Harbor API Test  ./tests/apitests/python/test_add_member_to_private_project.py
Test Case - Delete a Repository of a Certain Project Created by Normal User
    Harbor API Test  ./tests/apitests/python/test_del_repo.py
Test Case - Add a System Global Label to a Certain Tag
    Harbor API Test  ./tests/apitests/python/test_add_sys_label_to_tag.py
Test Case - Add a System Global Label to a Certain Tag
    Harbor API Test  ./tests/apitests/python/test_add_replication_rule.py
