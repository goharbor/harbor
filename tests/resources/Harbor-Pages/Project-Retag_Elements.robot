*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***

${retag_btn}  //clr-dg-action-bar//button[contains(.,'Retag')]
${project-name_xpath}  //*[@id='project-name']
${repo-name_xpath}  //*[@id='repo-name']
${tag-name_xpath}  //*[@id='tag-name']
${confirm_btn}  //button[contains(.,'CONFIRM')]
${target_image_name}  target-alpine
${target_tag_value}  3.2.10-target
${tag_value_xpath}  //clr-dg-row[contains(.,'${target_tag_value}')]
${image_tag}  3.2.10-alpine
${modal-dialog}  div.modal-dialog