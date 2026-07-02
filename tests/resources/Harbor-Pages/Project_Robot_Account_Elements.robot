*** Settings ***
Documentation  This resource provides any keywords related to the Harbor robot-account feature

*** Variables ***
${project_robot_account_tabpage}  xpath=//*[self::button or self::a][contains(., 'Robot Accounts')]
${project_robot_account_create_btn}  xpath=//project-detail/app-robot-account//button
${project_robot_account_token_input}  xpath=//app-robot-account//hbr-copy-input//input
${project_robot_account_name_xpath}  //view-token//div[contains(@class,'robot-name')]//span
${project_robot_account_create_name_input}  //input[@id='name']
${project_robot_account_create_finish_btn}  //button[text()='Finish']
${project_robot_account_create_sexpiration_type_btn}  //select[@id='expiration-type']
${project_robot_account_token_expiration_days}  //*[@id='robotTokenExpiration']
${project_robot_account_secret_input}  //input[@id='provided_secret']
${project_robot_account_secret_confirm_input}  //input[@id='confirm_secret']
${project_robot_account_secret_toggle_btn}  //clr-icon[contains(@class,'clr-input-icon-action')]
