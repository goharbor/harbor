<!--
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->
<div class="container">
	<div class="page-header" style="margin-top: -10px;">
		<h2>{{i18n .Lang "welcome"}} {{.Username}}</h2>
	</div>
	<div row="tabpanel">
		<ul class="nav nav-tabs" id="tabProject"> 
			<li role="presentation"><a href="#tabMyProject" aria-controls="tabProject" role="tab" data-toggle="tab">{{i18n .Lang "my_projects"}}</a></li>
			<li role="presentation"><a href="#tabMyProject" aria-controls="tabProject" role="tab" data-toggle="tab">{{i18n .Lang "public_projects"}}</a></li>
			<li role="presentation" id="liAdminOption" style="visibility: hidden;"><a href="#tabAdminOption" aria-controls="tabAdminOption" role="tab" data-toggle="tab">{{i18n .Lang "admin_options"}}</a></li>
		</ul>
		<!-- tab panes -->
		<div class="tab-content">
			<div role="tabpanel" class="tab-pane" id="tabMyProject" style="margin-top: 15px;">
			<form class="form-inline">
			    <label class="sr-only" for="txtProjectName">{{i18n .Lang "project_name"}}:</label>
				<div class="input-group">	
				    <div class="input-group-addon">{{i18n .Lang "project_name"}}:</div>    
					<input type="text" class="form-control" id="txtSearchProject">
					<span class="input-group-btn"> 
						<button id="btnSearch" type="button" class="btn btn-primary"><span class="glyphicon glyphicon-search"></span></button>
					</span>
				</div>
				<button type="button" class="btn btn-primary" data-toggle="modal" data-target="#dlgAddProject" id="btnAddProject">{{i18n .Lang "add_project"}}</button>					
			</form>
				<div class="table-responsive div-height">
					<table id="tblProject" class="table table-hover">
						<thead>
							<tr>
								<th width="35%">{{i18n .Lang "project_name"}}</th>
								<th width="45%">{{i18n .Lang "creation_time"}}</th>
				                                <th width="20%">{{i18n .Lang "publicity"}}</th> 
							</tr>
						</thead>
						<tbody>				
						</tbody>
					</table>
				</div>
			</div>
			<div role="tabpanel" class="tab-pane" id="tabAdminOption" style="visibility: hidden; margin-top: 15px;">
				   <form class="form-inline">
					   <label class="sr-only" for="txtProjectName">{{i18n .Lang "username"}}:</label>
					   <div class="input-group">
						    <div class="input-group-addon">{{i18n .Lang "username"}}:</div>
							<input type="text" class="form-control" id="txtSearchUsername">
							<span class="input-group-btn">
								<button id="btnSearchUsername" type="button" class="btn btn-primary"><span class="glyphicon glyphicon-search"></span></button>			
							</span>
						</div>
					</form>
					<div class="table-responsive div-height">
						<table id="tblUser" class="table table-hover">
							<thead>
								<tr>
									<th width="35%">{{i18n .Lang "username"}}</th>
									<th width="45%">{{i18n .Lang "email"}}</th>
									<th width="20%">{{i18n .Lang "system_admin"}}</th>
									<th></th>
								</tr>
							</thead>
							<tbody>
							</tbody>
						</table>
					</div>
				</div>
			</div>
		</div>
	</div>
	<div class="modal fade" id="dlgAddProject" tabindex="-1" role="dialog" aria-labelledby="Add Project" aria-hidden="true">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<a type="button" class="close" data-dismiss="modal" aria-label="Close" id="btnCancel">
						<span aria-hidden="true">&times;</span>
					</a>
					<h4 class="modal-title" id="dlgAddProjectTitle">{{i18n .Lang "add_project"}}</h4>
				</div>
				<div class="modal-body">
					<form role="form">
						<div class="alert alert-danger" role="alert" id="divErrMsg"></div>
						<div class="form-group has-feedback">
							<label for="projectName" class="control-label">{{i18n .Lang "project_name"}}:</label>
							<input type="text" class="form-control" id="projectName">
							<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
						</div>
                        <div class="checkbox">
                            <label>
                                <input type="checkbox" id="isPublic" checked=false> {{i18n .Lang "check_for_publicity"}}
                            </label>
                        </div>
					</form>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-primary" id="btnSave">{{i18n .Lang "button_save"}}</button>
					<button type="button" class="btn btn-default" data-dismiss="modal">{{i18n .Lang "button_cancel"}}</button>
				</div>
			</div>
		</div>
	</div>
</div>
<script src="static/resources/js/validate-options.js"></script>
<script src="static/resources/js/project.js"></script>
