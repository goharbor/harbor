*** Settings ***
Documentation  This resource contains any keywords related to using the Drone CI Build System

*** Keywords ***
Get State Of Drone Build
    [Arguments]  ${num}
    Return From Keyword If  '${num}' == '0'  local
    ${out}=  Run  drone build info vmware/vic ${num}
    ${lines}=  Split To Lines  ${out}
    [Return]  @{lines}[2]

Get Title of Drone Build
    [Arguments]  ${num}
    Return From Keyword If  '${num}' == '0'  local
    ${out}=  Run  drone build info vmware/vic ${num}
    ${lines}=  Split To Lines  ${out}
    [Return]  @{lines}[-1]
