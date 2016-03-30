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
	<ol class="breadcrumb" style="background: none;">
		<li><a href="/registry/project">{{i18n .Lang "project"}}</a></li>
		<li>{{.ProjectName}}</li>
	</ol>
	<div class="page-header"  style="margin-top: -10px;">
		<h2>{{.ProjectName}} </h2></h4>{{i18n .Lang "owner"}}: {{.OwnerName}}</h4>
	</div>
	<div row="tabpanel">
		<div class="row">
			<div class="col-md-2">
				<ul class="nav nav-pills nav-stacked" id="tabItemDetail">
					<li role="presentation"><a href="#tabRepoInfo" aria-controls="tabRepoInfo" role="tab" data-toggle="tab">{{i18n .Lang "repo"}}</a></li> 
					<li role="presentation" style="visibility: hidden;"><a href="#tabUserInfo" aria-controls="tabUserInfo" role="tab" data-toggle="tab">{{i18n .Lang "user"}}</a></li>
					<li role="presentation" style="visibility: hidden;"><a href="#tabOperationLog" aria-controls="tabOperationLog" role="tab" data-toggle="tab">{{i18n .Lang "logs"}}</a></li>
				</ul>
			</div>
		    <div class="col-md-10">
			<input type="hidden" id="projectId" value="{{.ProjectId}}">
			<input type="hidden" id="projectName" value="{{.ProjectName}}">
			<input type="hidden" id="userId" value="{{.UserId}}">
			<input type="hidden" id="ownerId" value="{{.OwnerId}}">
			<input type="hidden" id="roleId" value="{{.RoleId}}">
			<input type="hidden" id="harborRegUrl" value="{{.HarborRegUrl}}">
			<input type="hidden" id="public" value="{{.Public}}">
			<input type="hidden" id="repoName" value="{{.RepoName}}">
			<!-- tab panes -->
			<div class="tab-content">
				<div role="tabpanel" class="tab-pane" id="tabRepoInfo">
					<form class="form-inline">						
						<div class="form-group">
						    <label class="sr-only" for="txtRepoName">{{i18n .Lang "repo_name"}}:</label>
							<div class="input-group">							
								<div class="input-group-addon">{{i18n .Lang "repo_name"}}:</div>
								<input type="text" class="form-control" id="txtRepoName">
								<span class="input-group-btn"> 
									<button id="btnSearchRepo" type="button" class="btn btn-primary"><span class="glyphicon glyphicon-search"></span></button>
								</span>
							</div>
						</div>
					</form>
					<p>
					<div class="table-responsive div-height">
						<div class="alert alert-danger" role="alert" id="divErrMsg"><center></center></div>
						<div class="panel-group" id="accordionRepo" role="tablist" aria-multiselectable="true">
						</div>
					</div>
				</div>		
				<div role="tabpanel" class="tab-pane" id="tabUserInfo">
					<form class="form-inline">						
						<div class="form-group">
							<div class="input-group">							
								<label class="sr-only" for="txtSearchUser">{{i18n .Lang "username"}}:</label>
								<div class="input-group">	
								    <div class="input-group-addon">{{i18n .Lang "username"}}:</div>
									<input type="text" class="form-control" id="txtSearchUser">
									<span class="input-group-btn"> 
										<button id="btnSearchUser" type="button" class="btn btn-primary"><span class="glyphicon glyphicon-search"></span></button>
									</span>
								</div>
							</div>
						</div>
						<button type="button" class="btn btn-primary" data-toggle="modal" data-target="#dlgUser" id="btnAddUser">{{i18n .Lang "add_members"}}</button>
					</form>
					<p>
					<div class="table-responsive div-height">
						<table id="tblUser" class="table table-hover">
							<thead>
								<tr>
									<th>{{i18n .Lang "username"}}</th>
									<th>{{i18n .Lang "role"}}</th>
									<th>{{i18n .Lang "operation"}}</th>
								</tr>
							</thead>
							<tbody>
							</tbody>
						</table>
					</div>		
				</div>
				<div role="tabpanel" class="tab-pane" id="tabOperationLog">
					<form class="form-inline">
						<div class="form-group">
						    <label for="txtUserName" class="sr-only">{{i18n .Lang "username"}}:</label>
							<div class="input-group">
								<div class="input-group-addon">{{i18n .Lang "username"}}:</div>
								<input type="text" class="form-control" id="txtSearchUserName">
								<span class="input-group-btn"> 
									<button id="btnFilterLog" type="button" class="btn btn-primary" data-toggle="modal" data-target="#dlgSearch"><span class="glyphicon glyphicon-search"></span></button>
								</span>
							</div>
						</div>
						<div class="form-group">
							<div class="input-group">
								<button class="btn btn-link" type="button" data-toggle="collapse" data-target="#collapseAdvance" aria-expanded="false" aria-controls="collapseAdvance">{{i18n .Lang "advance"}}</button>
							</div>
						</div>
					<form>
					<p></p>
					<div class="collapse" id="collapseAdvance">
						<form class="form">
							<div class="form-group">
								<label for="txtUserName" class="sr-only">{{i18n .Lang "operation"}}:</label>
								<div class="input-group">
									<div class="input-group-addon">{{i18n .Lang "operation"}}:</div>	
								    <span class="input-group-addon" id="spnFilterOption">
										<input type="checkbox" name="chkAll" value="0"> {{i18n .Lang "all"}}
										<input type="checkbox" name="chkOperation" value="create"> Create
										<input type="checkbox" name="chkOperation" value="pull"> Pull
										<input type="checkbox" name="chkOperation" value="push"> Push
										<input type="checkbox" name="chkOperation" value="delete"> Delete
										<input type="checkbox" name="chkOperation" value="others"> {{i18n .Lang "others"}}:
										<input type="text" id="txtOthers" size="10">								
									</span>
								</div>
							</div>
							<p></p>
							<div class="form-group">
							    <label for="begindatepicker" class="sr-only">{{i18n .Lang "start_date"}}:</label>
								<div class="input-group">	
									<div class="input-group-addon">{{i18n .Lang "start_date"}}:</div>
					                  <div class="input-group date" id="datetimepicker1">
					                    <input type="text" class="form-control" id="begindatepicker" readonly="readonly">
					                    <span class="input-group-addon">
					                        <span class="glyphicon glyphicon-calendar"></span>
					                    </span>
					                  </div>
								</div>
								
							</div>
							<div class="form-group">
								<div class="input-group">	
									<div class="input-group-addon">{{i18n .Lang "end_date"}}:</div>
					                <div class="input-group date" id="datetimepicker2">
				                      <input type="text" class="form-control" id="enddatepicker" readonly="readonly">
				                      <span class="input-group-addon">
				                        <span class="glyphicon glyphicon-calendar"></span>
				                      </span>
				                    </div>
								</div>
								
							</div>			
						</form>	
					</div>				
					<div class="table-responsive div-height">
						<table id="tblAccessLog" class="table table-hover" >
							<thead>
								<tr>
									<th width="20%">{{i18n .Lang "username"}}</th>
									<th width="40%">{{i18n .Lang "repo_name"}}</th>
									<th width="20%">{{i18n .Lang "operation"}}</th>
									<th width="20%">{{i18n .Lang "timestamp"}}</th>
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
</div>
	<div class="modal fade" id="dlgUser" tabindex="-1" role="dialog" aria-labelledby="User" aria-hidden="true">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<button type="button" class="close" data-dismiss="modal" aria-label="Close">
						<span aria-hidden="true">&times;</span>
					</button>
					<h4 class="modal-title" id="dlgUserTitle"></h4>
				</div>
				<div class="modal-body">
					<form role="form">
						<input type="hidden" id="operationType" value="">
						<input type="hidden" id="editUserId" value="">
						<div class="form-group">
							<div class="input-group">
								<label for="txtUserName" class="input-group-addon">{{i18n .Lang "username"}}:</label>
								<input type="text" class="form-control" id="txtUserName">
							</div>
						</div>
						<div class="form-group">
							<label for="txtRole" class="control-label">{{i18n .Lang "role"}}:</label>
							<ul class="list-group" id="lstRole">
								<li class="list-group-item">
									<input type="radio" name="chooseRole" id="chkRole2" value="1">
									<label for="chkRole2" class="control-label">Project Admin</label>
								</li>
								<li class="list-group-item">
									<input type="radio" name="chooseRole" id="chkRole3" value="2">
									<label for="chkRole3" class="control-label">Developer</label>
								</li>
								<li class="list-group-item">
									<input type="radio" name="chooseRole" id="chkRole4" value="3">
									<label for="chkRole4" class="control-label">Guest</label>
								</li>
							</ul>
						</div>
					</form>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-primary" id="btnSave">{{i18n .Lang "button_save"}}</button>
					<button type="button" class="btn btn-default" data-dismiss="modal" id="btnCancel">{{i18n .Lang "button_cancel"}}</button>
				</div>
			</div>
		</div>
	</div>
</div>
<script src="static/resources/js/item-detail.js"></script>