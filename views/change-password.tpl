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
	<div class="col-sm-4"></div>
	<div class="col-sm-4">
		<div class="page-header">
				<h1>{{i18n .Lang "title_change_password"}}</h1>
		</div>
		<form class="form">
 		  <div class="alert alert-danger" role="alert" id="divErrMsg"></div>
		  <div class="form-group has-feedback">
		    <label for="OldPassword" class="control-label">{{i18n .Lang "old_password"}}</label>
		    <input type="password" class="form-control" id="OldPassword">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
		  </div>
		  <div class="form-group has-feedback">
		    <label for="Password" class="control-label">{{i18n .Lang "new_password"}}</label>
		    <input type="password" class="form-control" id="Password">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "password_description"}}</h6> 
		  </div>
		  <div class="form-group has-feedback">
		    <label for="ConfirmedPassword" class="control-label">{{i18n .Lang "confirm_password"}}</label>
		    <input type="password" class="form-control" id="ConfirmedPassword">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "password_description"}}</h6> 
		  </div>
		  <div class="form-group has-feedback">
		    <div class="text-center">
		      <button type="button" class="btn btn-default" id="btnSubmit">{{i18n .Lang "button_submit"}}</button>
		    </div>
		  </div>
		</form>
	</div>
	<div class="col-sm-4"></div>
</div>
<script src="static/resources/js/validate-options.js"></script>
<script src="static/resources/js/change-password.js"></script>