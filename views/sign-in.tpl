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
  <form class="form-signin form-horizontal">
	<div class="form-group">
    	<label for="Principal" class="col-md-4 control-label">{{i18n .Lang "username_email"}}</label>
		<div class="col-md-8">
    		<input type="text" id="Principal" class="form-control" placeholder="{{i18n .Lang "username_email"}}">
		</div>
	</div>
	<div class="form-group">
    	<label for="Password" class="col-md-4 control-label">{{i18n .Lang "password"}}</label>
		<div class="col-md-8">
    		<input type="password" id="Password" class="form-control" placeholder="{{i18n .Lang "password"}}">
		</div>
	</div>
    <button class="btn btn-lg btn-primary btn-block" type="button" id="btnPageSignIn">{{i18n .Lang "sign_in"}}</button>
	{{ if eq .AuthMode "db_auth" }}
	<div class="form-group">
	   <div class="col-md-12">
	      <button type="button" class="btn btn-link pull-right" id="btnForgot">{{i18n .Lang "forgot_password"}}</button>
	   </div>
	</div>
	{{ end }}
  </form>
</div>
<link href="static/resources/css/sign-in.css" type="text/css" rel="stylesheet">
<script src="static/resources/js/sign-in.js"></script>