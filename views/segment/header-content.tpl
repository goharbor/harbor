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
<input type="hidden" id="currentLanguage" value="{{.Lang}}">
<input type="hidden" id="isAdmin" value="{{.IsAdmin}}">
<nav class="navbar navbar-default" role="navigation" style="margin-bottom: 0;">
	<div class="navbar-header">
		<button aria-controls="navbar" aria-expanded="false" data-target="#navbar" data-toggle="collapse" class="navbar-toggle collapsed" type="button">
           <span class="sr-only">Toggle navigation</span>
           <span class="icon-bar"></span>
           <span class="icon-bar"></span>
        </button>
		<a class="navbar-brand" href="/"><img src="static/resources/image/Harbor_Logo_rec.png" height="40px" width="80px"/></a>
    </div>
	<div id="navbar" class="navbar-collapse collapse">
		<form class="navbar-form navbar-right">
			<div class="form-group">
				<div class="input-group">
				    <ul class="nav navbar-nav">
				    <li class="dropdown">
					  	<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false">
						   <span class="glyphicon glyphicon-globe"></span>
						     {{i18n .Lang "language"}}						   
						<span class="caret"></span></a>
				        <ul class="dropdown-menu">
						  <li><a href="/language?lang=en-US">{{i18n .Lang "language_en-US"}}</a></li>						
						  <li><a href="/language?lang=zh-CN">{{i18n .Lang "language_zh-CN"}}</a></li>
						  <li><a href="/language?lang=de-DE">{{i18n .Lang "language_de-DE"}}</a></li>
						</ul>
					</li>
			   	  </ul>
				</div>
			
			    <div class="input-group" >
				  <span class="input-group-addon"><span class="input-group glyphicon glyphicon-search"></span></span>
				  <input type="text" class="form-control" id="txtCommonSearch" size="50" placeholder="{{i18n .Lang "search_placeholder"}}">  	
				</div>
			</div>		
			{{ if .Username }}
			  <div class="input-group">
			    <ul class="nav navbar-nav">
			    <li class="dropdown">
				  	<a href="#" class="dropdown-toggle" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false"><span class="glyphicon glyphicon-user"></span> {{.Username}}<span class="caret"></span></a>
			        <ul class="dropdown-menu">
					    {{ if eq .AuthMode "db_auth" }}
						<li><a id="aChangePassword" href="/changePassword" target="_blank"><span class="glyphicon glyphicon-pencil"></span>&nbsp;&nbsp;{{i18n .Lang "change_password"}}</a></li>
						<li role="separator" class="divider"></li>
						{{ end }}
						{{ if eq .AuthMode "db_auth" }}
						  {{ if eq .IsAdmin true }}
						    <li><a id="aAddUser" href="/addUser" target="_blank"><span class="glyphicon glyphicon-plus"></span>&nbsp;&nbsp;{{i18n .Lang "add_user"}}</a></li>
					      {{ end }}
						{{ end}}
						<li><a id="aLogout" href="#"><span class="glyphicon glyphicon-log-in"></span>&nbsp;&nbsp;{{i18n .Lang "log_out"}}</a></li>
					</ul>
				</li>
			  	</ul>
			  </div>
			{{ else if eq .AuthMode "db_auth" }}
			  <div class="input-group">
	  		    &nbsp;<button type="button" class="btn btn-default" id="btnSignIn">{{i18n .Lang "sign_in"}}</button>
				{{ if eq .SelfRegistration true }}
				&nbsp;<button type="button" class="btn btn-success" id="btnSignUp">{{i18n .Lang "sign_up"}}</button>
				{{ end }}
			  </div>
		    {{ else }}
			  <div class="input-group">
	  		    &nbsp;<button type="button" class="btn btn-default" id="btnSignIn">{{i18n .Lang "sign_in"}}</button>
			  </div>
			{{ end }}
		</form>	 
	</div>		 
</nav>
