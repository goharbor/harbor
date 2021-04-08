*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***

${retag_btn}  //clr-dg-action-bar//button[contains(.,'Retag')]
${copy_project_name_xpath}  //*[@id='project-name']
${copy_repo_name_xpath}  //*[@id='repo-name']
${tag_name_xpath}  //*[@id='tag-name']
${confirm_btn}  //button[contains(.,'CONFIRM')]
${target_image_name}  target-alpine
${image_tag}  3.2.10-alpine
${tag_value_xpath}  //clr-dg-row[contains(.,'${image_tag}')]
${modal-dialog}  div.modal-dialog
