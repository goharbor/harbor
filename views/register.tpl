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
				{{ if eq .IsAdmin true }}
				<h1>{{i18n .Lang "add_user" }}</h1>
				{{ else }}
				<h1>{{i18n .Lang "registration"}}</h1>
				{{ end }}
		</div>
		<form class="form">
		  <div class="alert alert-danger" role="alert" id="divErrMsg"></div>
		  <div class="form-group has-feedback">
		    <label for="username" class="control-label">{{i18n .Lang "username"}}</label>
			<p style="display:inline; color: red; font-size: 12pt;">*</p>
		    <input type="text" class="form-control" id="Username">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "username_description"}}</h6>
		  </div>
		  <div class="form-group has-feedback">
		    <label for="Email" class="control-label">{{i18n .Lang "email"}}</label>
			<p style="display:inline; color: red; font-size: 12pt;">*</p>
		    <input type="email" class="form-control" id="Email">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "email_description"}}</h6>
		  </div>
		  <div class="form-group has-feedback">
		    <label for="Realname" class="control-label">{{i18n .Lang "full_name"}}</label>
			<p style="display:inline; color: red; font-size: 12pt;">*</p>
		    <input type="text" class="form-control" id="Realname">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "full_name_description"}}</h6>
		  </div>
		  <div class="form-group has-feedback">
		    <label for="Password" class="control-label">{{i18n .Lang "password"}}</label>
			<p style="display:inline; color: red; font-size: 12pt;">*</p>
		    <input type="password" class="form-control" id="Password">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "password_description"}}</h6> 
		  </div>
		  <div class="form-group has-feedback">
		    <label for="ConfirmedPassword" class="control-label">{{i18n .Lang "confirm_password"}}</label>
			<p style="display:inline; color: red; font-size: 12pt;">*</p>
		    <input type="password" class="form-control" id="ConfirmedPassword">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
			<h6>{{i18n .Lang "password_description"}}</h6> 
		  </div>
	 	  <div class="form-group has-feedback">
		    <label for="Comment" class="control-label">{{i18n .Lang "note_to_the_admin"}}</label>
		    <input type="text" class="form-control" id="Comment">
			<span class="glyphicon form-control-feedback" aria-hidden="true"></span>
		  </div>
		  <div class="form-group has-feedback">
		    <div class="text-center">
		      <button type="button" class="btn btn-default" id="btnPageSignUp">
				{{ if eq .IsAdmin true }}
			        {{i18n .Lang "add_user" }}
			    {{ else }} 
					{{i18n .Lang "sign_up"}}
				{{ end }}
			  </button>
		    </div>
		  </div>
		</form>
	</div>
	<div class="col-sm-4"></div>
</div>
<script src="static/resources/js/validate-options.js"></script>
<script src="static/resources/js/register.js"></script>