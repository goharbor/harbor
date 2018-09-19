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
Test Case - Create Project
    ${project_id}  ${project_name} =  Create Project
    Log To Console  ${project_name}

Test Case - Push Image
    Sleep  1
    Docker Pull  hello-world:latest
    ${project_id}  ${project_name} =  Create Project
    Log To Console  ${project_name}
    Docker Login  ${SERVER}  admin  Harbor12345
    Docker Tag  hello-world:latest  ${SERVER}/${project_name}/hello-world:1.0
    Docker Push  ${SERVER}/${project_name}/hello-world:1.0
