*** Settings ***
Documentation  This resource provides any keywords related to the Harbor robot-account feature

*** Variables ***
${project_robot_account_tabpage}  xpath=//project-detail//a[contains(.,'Robot Accounts')]
${project_robot_account_create_btn}  xpath=//project-detail/app-robot-account//button
${project_robot_account_token_input}  xpath=//app-robot-account//hbr-copy-input//input
${project_robot_account_create_name_input}  //input[@id='robot_name']
${project_robot_account_create_save_btn}  //add-robot//button[contains(.,'SAVE')]
