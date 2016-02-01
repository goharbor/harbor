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
<input type="hidden" id="queryParam" value="{{.QueryParam}}">
<div class="container">
	<ol class="breadcrumb">
		<li><a href="/">{{i18n .Lang "home"}}</a></li>
		<li>{{i18n .Lang "search"}}</li>
	</ol>
    <div class="panel panel-default">
	     <div class="panel-heading" id="panelCommonSearchProjectsHeader">{{i18n .Lang "projects"}}</div>
		 <div class="panel-body" id="panelCommonSearchProjectsBody">
		 </div>
	</div>
	<div class="panel panel-default">
	     <div class="panel-heading" id="panelCommonSearchRepositoriesHeader">{{i18n .Lang "repositories"}}</div>
		 <div class="panel-body" id="panelCommonSearchRepositoriesBody">
		 </div>
	</div>
</div>
<script src="static/resources/js/search.js"></script>