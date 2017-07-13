# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***
${create_project_button_css}  .btn
${project_name_xpath}  //*[@id="create_project_name"]
${project_public_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/div/project/div/div/div[2]/div[1]/create-project/clr-modal/div/div[1]/div/div[1]/div/div[2]/form/section/div[2]/div/label
${project_save_css}  html body.no-scrolling harbor-app harbor-shell clr-main-container.main-container div.content-container div.content-area.content-area-override project div.row div.col-lg-12.col-md-12.col-sm-12.col-xs-12 div.row.flex-items-xs-between div.option-left create-project clr-modal div.modal div.modal-dialog div.modal-content div.modal-footer button.btn.btn-primary
${log_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[2]
${projects_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[1]
${replication_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/div/project-detail/nav/ul/li[4]/a
