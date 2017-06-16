*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Bundle

*** Test Cases ***
Distro Harbor Offline
    Package Harbor Offline
