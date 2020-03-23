*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/APITest-Util.robot
Library  OperatingSystem
Library  String
Library  Collections
Library  requests
Library  Process
Default Tags  API

*** Test Cases ***
Test Case - LDAP Group Admin Role
    Harbor API Test  ./tests/apitests/python/test_ldap_admin_role.py

Test Case - LDAP Group User Group
    Harbor API Test  ./tests/apitests/python/test_user_group.py

Test Case - Run LDAP Group Related API Test
    Harbor API Test  ./tests/apitests/python/test_assign_role_to_ldap_group.py